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

const (
	testTemplateFrontend = "frontend"
	testEnvContent       = "NODE_ENV=development\nAPI_URL=http://localhost:3000"
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
	// Test ListSchemas (registered template schemas)
	client := New()
	schemas := client.ListSchemas()

	// Should return a slice (may be empty in test environment)
	if schemas == nil {
		t.Error("Expected non-nil schemas slice")
	}

	// Test ListTemplateTypes (built-in template types)
	templateTypes := client.ListTemplateTypes()

	// Should contain the registered template types
	expectedTypes := map[string]bool{testTemplateFrontend: true, "go-api": true, "fullstack": true}
	if len(templateTypes) != len(expectedTypes) {
		t.Errorf("Expected %d template types, got %d", len(expectedTypes), len(templateTypes))
	}

	for _, templateName := range templateTypes {
		if !expectedTypes[templateName] {
			t.Errorf("Unexpected template type: %q", templateName)
		}
		delete(expectedTypes, templateName)
	}

	if len(expectedTypes) > 0 {
		t.Errorf("Missing template types: %v", expectedTypes)
	}
}

func TestExtractWithMockTemplate(t *testing.T) {
	mockClient := createMockClient()

	// Test extraction error with invalid template type
	_, err := mockClient.Extract(context.Background(), ExtractOptions{
		SourceDir: "/tmp",
		Type:      "invalid-type",
	})
	if err == nil {
		t.Error("Expected error for invalid template type")
	}
}

func TestGetTemplateInfo(t *testing.T) {
	client := New()

	tests := []struct {
		name         string
		templateType string
		wantErr      bool
		wantName     string
		wantVarCount int
	}{
		{
			name:         "valid template type",
			templateType: testTemplateFrontend,
			wantErr:      false,
			wantName:     testTemplateFrontend,
			wantVarCount: 4, // ProjectName, GitHubRepo, Author, Description
		},
		{
			name:         "invalid template type",
			templateType: "nonexistent",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.GetTemplateTypeInfo(tt.templateType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetTemplateTypeInfo() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetTemplateTypeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if info == nil {
				t.Fatal("GetTemplateTypeInfo() returned nil info")
			}

			if info.Name != tt.wantName {
				t.Errorf("GetTemplateTypeInfo() name = %v, want %v", info.Name, tt.wantName)
			}

			if len(info.Variables) != tt.wantVarCount {
				t.Errorf("GetTemplateTypeInfo() variables count = %v, want %v", len(info.Variables), tt.wantVarCount)
			}

			// Check that variables have expected structure
			for name, variable := range info.Variables {
				if variable.Type == "" {
					t.Errorf("Variable %s has empty type", name)
				}
			}
		})
	}
}

func TestGetTemplateVariables(t *testing.T) {
	client := New()

	tests := []struct {
		name         string
		templateType string
		wantErr      bool
		wantVarCount int
	}{
		{
			name:         "valid template type",
			templateType: testTemplateFrontend,
			wantErr:      false,
			wantVarCount: 4, // ProjectName, GitHubRepo, Author, Description
		},
		{
			name:         "invalid template type",
			templateType: "nonexistent",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeInfo, err := client.GetTemplateTypeInfo(tt.templateType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetTemplateTypeInfo() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetTemplateTypeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			variables := typeInfo.Variables
			if len(variables) != tt.wantVarCount {
				t.Errorf("GetTemplateTypeInfo().Variables count = %v, want %v", len(variables), tt.wantVarCount)
			}

			// Check that variables have expected structure
			for name, variable := range variables {
				if variable.Type == "" {
					t.Errorf("Variable %s has empty type", name)
				}
			}
		})
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

func TestGetTemplateEnvConfig(t *testing.T) {
	client := New()

	// Create a test template with environment configuration
	testTemplate := &core.TemplateSchema{
		Name:        "test-template",
		Type:        "test",
		Version:     "1.0.0",
		Description: "Test template with env config",
		Variables: map[string]core.Variable{
			"ProjectName": {Type: "string", Required: true},
		},
		Files: []core.FileSpec{
			{Path: "test.txt", Content: "test", Template: false},
		},
		EnvConfig: []core.EnvVariable{
			{Name: "DB_HOST", Description: "Database host", Example: "localhost"},
			{Name: "DB_PORT", Description: "Database port", Example: "5432"},
			{Name: "API_KEY", Description: "API key for external service", Example: "your-api-key"},
		},
	}

	// Register the test template
	client.templates["test-template"] = testTemplate

	tests := []struct {
		name         string
		templateName string
		wantErr      bool
		expectedLen  int
	}{
		{
			name:         "valid template with env config",
			templateName: "test-template",
			wantErr:      false,
			expectedLen:  3,
		},
		{
			name:         "non-existent template",
			templateName: "non-existent",
			wantErr:      true,
			expectedLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envConfig, err := client.GetSchemaEnvConfig(tt.templateName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSchemaEnvConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(envConfig) != tt.expectedLen {
					t.Errorf("GetSchemaEnvConfig() returned %d env vars, expected %d", len(envConfig), tt.expectedLen)
				}

				// Check specific environment variables
				expectedVars := map[string]struct {
					description string
					example     string
				}{
					"DB_HOST": {"Database host", "localhost"},
					"DB_PORT": {"Database port", "5432"},
					"API_KEY": {"API key for external service", "your-api-key"},
				}

				for _, envVar := range envConfig {
					if expected, exists := expectedVars[envVar.Name]; exists {
						if envVar.Description != expected.description {
							t.Errorf("EnvVar %s description = %v, expected %v", envVar.Name, envVar.Description, expected.description)
						}
						if envVar.Example != expected.example {
							t.Errorf("EnvVar %s example = %v, expected %v", envVar.Name, envVar.Example, expected.example)
						}
					} else {
						t.Errorf("Unexpected environment variable: %s", envVar.Name)
					}
				}
			}
		})
	}
}

// TestTemplateTypesAPI tests the Template Types API (built-in extractors)
func TestTemplateTypesAPI(t *testing.T) {
	client := New()

	t.Run("ListTemplateTypes", func(t *testing.T) {
		types := client.ListTemplateTypes()
		if len(types) == 0 {
			t.Fatal("ListTemplateTypes returned empty list")
		}

		// Verify expected template types are registered
		expectedTypes := map[string]bool{
			testTemplateFrontend: false,
			"go-api":             false,
			"fullstack":          false,
		}

		for _, templateType := range types {
			if _, exists := expectedTypes[templateType]; exists {
				expectedTypes[templateType] = true
			}
		}

		for templateType, found := range expectedTypes {
			if !found {
				t.Errorf("Expected template type %s not found in list", templateType)
			}
		}
	})

	t.Run("GetTemplateTypeInfo", func(t *testing.T) {
		typeInfo, err := client.GetTemplateTypeInfo(testTemplateFrontend)
		if err != nil {
			t.Fatalf("GetTemplateTypeInfo failed: %v", err)
		}

		if typeInfo.Name != testTemplateFrontend {
			t.Errorf("Expected name '%s', got %s", testTemplateFrontend, typeInfo.Name)
		}
		if typeInfo.Description == "" {
			t.Error("Description should not be empty")
		}
		if typeInfo.Variables == nil {
			t.Error("Variables should not be nil")
		}
	})

	t.Run("GetTemplateTypeInfo_NotFound", func(t *testing.T) {
		_, err := client.GetTemplateTypeInfo("nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent template type")
		}
	})

	t.Run("ExtractSchema", func(t *testing.T) {
		// Create a temporary source directory
		tempDir := t.TempDir()

		// Create some basic files that the frontend template would expect
		err := os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(`{"name": "test-app"}`), 0o644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Create .env.example file
		err = os.WriteFile(filepath.Join(tempDir, ".env.example"), []byte(testEnvContent), 0o644)
		if err != nil {
			t.Fatalf("Failed to create .env.example: %v", err)
		}

		schema, err := client.ExtractSchema(testTemplateFrontend, tempDir)
		if err != nil {
			t.Fatalf("ExtractSchema failed: %v", err)
		}

		if schema.Name == "" {
			t.Error("Schema name should not be empty")
		}
		if schema.Type != testTemplateFrontend {
			t.Errorf("Expected type '%s', got %s", testTemplateFrontend, schema.Type)
		}
	})
}

// TestTemplateSchemasAPI tests the Template Schemas API (registered data)
func TestTemplateSchemasAPI(t *testing.T) {
	client := New()

	// Create a test template schema file
	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test-template.json")

	testSchema := &core.TemplateSchema{
		Name:        "test-template",
		Type:        testTemplateFrontend,
		Version:     "1.0.0",
		Description: "Test template schema",
		Variables: map[string]core.Variable{
			"project_name": {
				Type:        "string",
				Description: "Name of the project",
				Default:     "my-app",
				Required:    true,
			},
		},
		EnvConfig: []core.EnvVariable{
			{
				Name:        "NODE_ENV",
				Description: "Node environment",
				Example:     "development",
			},
		},
		Files: []core.FileSpec{
			{
				Path:     "package.json",
				Template: true,
				Content:  `{"name": "{{.project_name}}"}`,
				Size:     35,
			},
		},
	}

	schemaJSON, err := json.MarshalIndent(testSchema, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test schema: %v", err)
	}

	err = os.WriteFile(schemaFile, schemaJSON, 0o644)
	if err != nil {
		t.Fatalf("Failed to write test schema file: %v", err)
	}

	t.Run("RegisterSchema", func(t *testing.T) {
		err := client.RegisterSchema(schemaFile)
		if err != nil {
			t.Fatalf("RegisterSchema failed: %v", err)
		}
	})

	t.Run("ListSchemas", func(t *testing.T) {
		schemas := client.ListSchemas()
		if len(schemas) == 0 {
			t.Fatal("ListSchemas returned empty list after registration")
		}

		found := false
		for _, schema := range schemas {
			if schema == "test-template" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Registered schema not found in list")
		}
	})

	t.Run("GetSchemaInfo", func(t *testing.T) {
		schemaInfo, err := client.GetSchemaInfo("test-template")
		if err != nil {
			t.Fatalf("GetSchemaInfo failed: %v", err)
		}

		if schemaInfo.Name != "test-template" {
			t.Errorf("Expected name 'test-template', got %s", schemaInfo.Name)
		}
		if schemaInfo.Type != testTemplateFrontend {
			t.Errorf("Expected type '%s', got %s", testTemplateFrontend, schemaInfo.Type)
		}
		if schemaInfo.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got %s", schemaInfo.Version)
		}
		if schemaInfo.FileCount != 1 {
			t.Errorf("Expected file count 1, got %d", schemaInfo.FileCount)
		}
		if schemaInfo.EnvVarCount != 1 {
			t.Errorf("Expected env var count 1, got %d", schemaInfo.EnvVarCount)
		}
	})

	t.Run("GetSchemaEnvConfig", func(t *testing.T) {
		envVars, err := client.GetSchemaEnvConfig("test-template")
		if err != nil {
			t.Fatalf("GetSchemaEnvConfig failed: %v", err)
		}

		if len(envVars) != 1 {
			t.Fatalf("Expected 1 env var, got %d", len(envVars))
		}

		envVar := envVars[0]
		if envVar.Name != "NODE_ENV" {
			t.Errorf("Expected env var name 'NODE_ENV', got %s", envVar.Name)
		}
		if envVar.Description != "Node environment" {
			t.Errorf("Expected description 'Node environment', got %s", envVar.Description)
		}
		if envVar.Example != "development" {
			t.Errorf("Expected example 'development', got %s", envVar.Example)
		}
	})

	t.Run("GenerateFromSchema", func(t *testing.T) {
		outputDir := t.TempDir()

		variables := Variables{
			ProjectName: "test-project",
			GitHubRepo:  "user/test-repo",
			OutputDir:   outputDir,
		}

		err := client.GenerateFromSchema(context.Background(), "test-template", variables)
		if err != nil {
			t.Fatalf("GenerateFromSchema failed: %v", err)
		}

		// Verify the generated file exists
		generatedFile := filepath.Join(outputDir, "package.json")
		if _, err := os.Stat(generatedFile); os.IsNotExist(err) {
			t.Error("Expected generated file does not exist")
		}
	})
}

// TestTypeAliases verifies that type aliases work correctly
func TestTypeAliases(t *testing.T) {
	client := New()

	// Create a test schema file to verify type compatibility
	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "alias-test.json")

	// Use core types directly to create schema
	coreSchema := &core.TemplateSchema{
		Name:        "alias-test",
		Type:        testTemplateFrontend,
		Version:     "1.0.0",
		Description: "Test schema for type aliases",
		Variables: map[string]core.Variable{
			"test": {
				Type:        "string",
				Description: "Test variable",
				Default:     "value",
				Required:    true,
			},
		},
		EnvConfig: []core.EnvVariable{
			{
				Name:        "TEST_VAR",
				Description: "Test environment variable",
				Example:     "test-value",
			},
		},
		Files: []core.FileSpec{
			{
				Path:     "test.txt",
				Template: true,
				Content:  "{{.test}}",
				Size:     9,
			},
		},
	}

	schemaJSON, err := json.MarshalIndent(coreSchema, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal core schema: %v", err)
	}

	err = os.WriteFile(schemaFile, schemaJSON, 0o644)
	if err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Register the schema
	err = client.RegisterSchema(schemaFile)
	if err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	t.Run("VariableTypeAlias", func(t *testing.T) {
		schemaInfo, err := client.GetSchemaInfo("alias-test")
		if err != nil {
			t.Fatalf("GetSchemaInfo failed: %v", err)
		}

		// schemaInfo.Variables should be directly usable as Variable (alias)
		sdkVar := schemaInfo.Variables["test"]
		if sdkVar.Description != "Test variable" {
			t.Errorf("Type alias failed for Variable: expected 'Test variable', got %s", sdkVar.Description)
		}
	})

	t.Run("EnvVariableTypeAlias", func(t *testing.T) {
		envVars, err := client.GetSchemaEnvConfig("alias-test")
		if err != nil {
			t.Fatalf("GetSchemaEnvConfig failed: %v", err)
		}

		if len(envVars) != 1 {
			t.Fatalf("Expected 1 env var, got %d", len(envVars))
		}

		// envVars[0] should be directly usable as EnvVariable (alias)
		sdkEnvVar := envVars[0]
		if sdkEnvVar.Name != "TEST_VAR" {
			t.Errorf("Type alias failed for EnvVariable: expected 'TEST_VAR', got %s", sdkEnvVar.Name)
		}
	})

	t.Run("TemplateSchemaTypeAlias", func(t *testing.T) {
		// Extract should return a TemplateSchema that's directly compatible
		tempSourceDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tempSourceDir, "package.json"), []byte(`{"name": "test"}`), 0o644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		schema, err := client.ExtractSchema(testTemplateFrontend, tempSourceDir)
		if err != nil {
			t.Fatalf("ExtractSchema failed: %v", err)
		}

		// schema should be directly usable as TemplateSchema (alias)
		sdkSchema := *schema
		if sdkSchema.Type != testTemplateFrontend {
			t.Errorf("Type alias failed for TemplateSchema: expected '%s', got %s", testTemplateFrontend, sdkSchema.Type)
		}
	})
}

func TestGetTemplateEnvConfigEmptyConfig(t *testing.T) {
	client := New()

	// Create a test template without environment configuration
	testTemplate := &core.TemplateSchema{
		Name:        "empty-env-template",
		Type:        "test",
		Version:     "1.0.0",
		Description: "Test template without env config",
		Variables: map[string]core.Variable{
			"ProjectName": {Type: "string", Required: true},
		},
		Files: []core.FileSpec{
			{Path: "test.txt", Content: "test", Size: 4},
		},
		EnvConfig: []core.EnvVariable{}, // Empty env config
	}

	client.templates["empty-env-template"] = testTemplate

	envConfig, err := client.GetSchemaEnvConfig("empty-env-template")
	if err != nil {
		t.Errorf("GetSchemaEnvConfig() unexpected error = %v", err)
	}

	if len(envConfig) != 0 {
		t.Errorf("GetSchemaEnvConfig() returned %d env vars, expected 0", len(envConfig))
	}
}
