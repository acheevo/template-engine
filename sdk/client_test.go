package sdk

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/acheevo/template-engine/internal/core"
	_ "github.com/acheevo/template-engine/internal/templates" // Register template types
)

func TestNew(t *testing.T) {
	client := New()
	if client == nil {
		t.Fatal("New() returned nil client")
	}
	if client.templates == nil {
		t.Fatal("Client templates map is nil")
	}
}

func TestValidateGenerateOptions(t *testing.T) {
	client := New()

	tests := []struct {
		name    string
		opts    GenerateOptions
		wantErr bool
		errType ErrorType
	}{
		{
			name: "valid options",
			opts: GenerateOptions{
				Template:    "frontend",
				ProjectName: "test-project",
				GitHubRepo:  "user/test-repo",
				OutputDir:   "./test-output",
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			opts: GenerateOptions{
				Template:   "frontend",
				GitHubRepo: "user/test-repo",
			},
			wantErr: true,
			errType: ErrorTypeValidation,
		},
		{
			name: "missing github repo",
			opts: GenerateOptions{
				Template:    "frontend",
				ProjectName: "test-project",
			},
			wantErr: true,
			errType: ErrorTypeValidation,
		},
		{
			name: "missing template",
			opts: GenerateOptions{
				ProjectName: "test-project",
				GitHubRepo:  "user/test-repo",
			},
			wantErr: true,
			errType: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateGenerateOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGenerateOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if sdkErr, ok := err.(*SDKError); ok {
					if sdkErr.Type != tt.errType {
						t.Errorf("Expected error type %v, got %v", tt.errType, sdkErr.Type)
					}
				} else {
					t.Errorf("Expected SDKError, got %T", err)
				}
			}
		})
	}
}

func TestValidateExtractOptions(t *testing.T) {
	client := New()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sdk-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		opts    ExtractOptions
		wantErr bool
		errType ErrorType
	}{
		{
			name: "valid options",
			opts: ExtractOptions{
				SourceDir: tempDir,
				Type:      "frontend",
			},
			wantErr: false,
		},
		{
			name: "missing source dir",
			opts: ExtractOptions{
				Type: "frontend",
			},
			wantErr: true,
			errType: ErrorTypeValidation,
		},
		{
			name: "missing type",
			opts: ExtractOptions{
				SourceDir: tempDir,
			},
			wantErr: true,
			errType: ErrorTypeValidation,
		},
		{
			name: "non-existent directory",
			opts: ExtractOptions{
				SourceDir: "/path/that/does/not/exist",
				Type:      "frontend",
			},
			wantErr: true,
			errType: ErrorTypeFileSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateExtractOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExtractOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if sdkErr, ok := err.(*SDKError); ok {
					if sdkErr.Type != tt.errType {
						t.Errorf("Expected error type %v, got %v", tt.errType, sdkErr.Type)
					}
				} else {
					t.Errorf("Expected SDKError, got %T", err)
				}
			}
		})
	}
}

func TestValidateVariables(t *testing.T) {
	client := New()

	tests := []struct {
		name    string
		vars    Variables
		wantErr bool
		errType ErrorType
	}{
		{
			name: "valid variables",
			vars: Variables{
				ProjectName: "test-project",
				GitHubRepo:  "user/test-repo",
				OutputDir:   "./test-output",
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			vars: Variables{
				GitHubRepo: "user/test-repo",
				OutputDir:  "./test-output",
			},
			wantErr: true,
			errType: ErrorTypeValidation,
		},
		{
			name: "missing github repo",
			vars: Variables{
				ProjectName: "test-project",
				OutputDir:   "./test-output",
			},
			wantErr: true,
			errType: ErrorTypeValidation,
		},
		{
			name: "missing output dir",
			vars: Variables{
				ProjectName: "test-project",
				GitHubRepo:  "user/test-repo",
			},
			wantErr: true,
			errType: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateVariables(tt.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if sdkErr, ok := err.(*SDKError); ok {
					if sdkErr.Type != tt.errType {
						t.Errorf("Expected error type %v, got %v", tt.errType, sdkErr.Type)
					}
				} else {
					t.Errorf("Expected SDKError, got %T", err)
				}
			}
		})
	}
}

func TestGenerateFromFile(t *testing.T) {
	client := New()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sdk-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template schema
	schema := &core.TemplateSchema{
		Name:        "test-template",
		Type:        "frontend",
		Version:     "1.0.0",
		Description: "Test template",
		Variables: map[string]core.Variable{
			"ProjectName": {Type: "string", Required: true},
			"GitHubRepo":  {Type: "string", Required: true},
		},
		Files: []core.FileSpec{
			{
				Path:     "README.md",
				Template: true,
				Content:  "# {{.ProjectName}}\n\nRepository: {{.GitHubRepo}}",
				Size:     50,
			},
			{
				Path:     "package.json",
				Template: true,
				Content:  `{"name": "{{.ProjectName}}", "repository": "{{.GitHubRepo}}"}`,
				Size:     60,
			},
		},
	}

	// Write schema to a temporary file
	schemaFile := filepath.Join(tempDir, "test-template.json")
	schemaData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(schemaFile, schemaData, 0o644); err != nil {
		t.Fatal(err)
	}

	outputDir := filepath.Join(tempDir, "output")

	tests := []struct {
		name         string
		templateFile string
		variables    Variables
		wantErr      bool
		errType      ErrorType
	}{
		{
			name:         "valid generation",
			templateFile: schemaFile,
			variables: Variables{
				ProjectName: "test-project",
				GitHubRepo:  "user/test-repo",
				OutputDir:   outputDir,
			},
			wantErr: false,
		},
		{
			name:         "non-existent template file",
			templateFile: "/path/that/does/not/exist.json",
			variables: Variables{
				ProjectName: "test-project",
				GitHubRepo:  "user/test-repo",
				OutputDir:   outputDir,
			},
			wantErr: true,
			errType: ErrorTypeFileSystem,
		},
		{
			name:         "invalid variables",
			templateFile: schemaFile,
			variables: Variables{
				GitHubRepo: "user/test-repo",
				OutputDir:  outputDir,
			},
			wantErr: true,
			errType: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.GenerateFromFile(context.Background(), tt.templateFile, tt.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if sdkErr, ok := err.(*SDKError); ok {
					if sdkErr.Type != tt.errType {
						t.Errorf("Expected error type %v, got %v", tt.errType, sdkErr.Type)
					}
				} else {
					t.Errorf("Expected SDKError, got %T", err)
				}
			}

			// If generation was successful, check that files were created
			if !tt.wantErr {
				readmeFile := filepath.Join(outputDir, "README.md")
				if _, err := os.Stat(readmeFile); os.IsNotExist(err) {
					t.Errorf("Expected README.md to be created at %s", readmeFile)
				}

				packageFile := filepath.Join(outputDir, "package.json")
				if _, err := os.Stat(packageFile); os.IsNotExist(err) {
					t.Errorf("Expected package.json to be created at %s", packageFile)
				}

				// Check content was templated correctly
				readmeContent, err := os.ReadFile(readmeFile)
				if err != nil {
					t.Errorf("Failed to read generated README.md: %v", err)
				}
				expectedContent := "# test-project\n\nRepository: user/test-repo"
				if string(readmeContent) != expectedContent {
					t.Errorf("README.md content mismatch.\nExpected: %q\nGot: %q", expectedContent, string(readmeContent))
				}
			}

			// Clean up output directory for next test
			os.RemoveAll(outputDir)
		})
	}
}

func TestListTemplates(t *testing.T) {
	// Test with empty client
	client := New()
	templates := client.ListTemplates()

	// Should return a slice (may be empty in test environment)
	if templates == nil {
		t.Error("Expected non-nil templates slice")
	}

	// Test with mock client
	mockClient := createMockClient()
	mockTemplates := mockClient.ListTemplates()

	if len(mockTemplates) != 2 {
		t.Errorf("Expected 2 mock templates, got %d", len(mockTemplates))
	}

	expectedTypes := map[string]bool{"mock-frontend": true, "mock-api": true}
	for _, templateType := range mockTemplates {
		if !expectedTypes[templateType] {
			t.Errorf("Unexpected template type: %q", templateType)
		}
		delete(expectedTypes, templateType)
	}

	if len(expectedTypes) > 0 {
		t.Errorf("Missing template types: %v", expectedTypes)
	}
}

func TestExtractWithMockTemplate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "extract-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	mockClient := createMockClient()

	schema, err := mockClient.Extract(context.Background(), ExtractOptions{
		SourceDir: tempDir,
		Type:      "mock-frontend",
	})
	if err != nil {
		t.Errorf("Extract() error = %v", err)
	}

	if schema == nil {
		t.Fatal("Extract() returned nil schema")
	}

	if schema.Name != "mock-template" {
		t.Errorf("Expected schema name 'mock-template', got %q", schema.Name)
	}

	if schema.Type != "mock-frontend" {
		t.Errorf("Expected schema type 'mock-frontend', got %q", schema.Type)
	}

	if len(schema.Files) != 1 {
		t.Errorf("Expected 1 file in schema, got %d", len(schema.Files))
	}
}

func TestValidate(t *testing.T) {
	client := New()

	tests := []struct {
		name    string
		schema  *core.TemplateSchema
		wantErr bool
	}{
		{
			name: "valid schema",
			schema: &core.TemplateSchema{
				Name:        "test-template",
				Type:        "frontend",
				Version:     "1.0.0",
				Description: "Test template",
				Variables: map[string]core.Variable{
					"ProjectName": {Type: "string", Required: true},
				},
				Files: []core.FileSpec{
					{
						Path:    "README.md",
						Content: "# Test",
						Size:    6,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			schema: &core.TemplateSchema{
				Type:        "frontend",
				Version:     "1.0.0",
				Description: "Test template",
				Variables: map[string]core.Variable{
					"ProjectName": {Type: "string", Required: true},
				},
				Files: []core.FileSpec{
					{
						Path:    "README.md",
						Content: "# Test",
						Size:    6,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing files",
			schema: &core.TemplateSchema{
				Name:        "test-template",
				Type:        "frontend",
				Version:     "1.0.0",
				Description: "Test template",
				Variables: map[string]core.Variable{
					"ProjectName": {Type: "string", Required: true},
				},
				Files: []core.FileSpec{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Validate(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
