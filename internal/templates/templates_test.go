package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFrontendTemplateExtractWithEnvExample(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "frontend-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal frontend project structure with .env.example
	projectFiles := map[string]string{
		"package.json": `{
  "name": "test-frontend",
  "version": "1.0.0"
}`,
		"src/App.tsx": `export default function App() {
  return <div>Hello World</div>;
}`,
		".env.example": `# Application name displayed in UI
APP_NAME="Test Frontend"
# Port for development server
PORT=3000
# API base URL
API_BASE_URL=http://localhost:8000/api`,
	}

	// Create the files
	for path, content := range projectFiles {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Test extraction
	frontend := &FrontendTemplate{}
	schema, err := frontend.Extract(tempDir)
	if err != nil {
		t.Fatalf("Failed to extract frontend template: %v", err)
	}

	// Verify basic schema properties
	if schema.Type != "frontend" {
		t.Errorf("Expected schema type 'frontend', got '%s'", schema.Type)
	}

	// Verify environment configuration was extracted
	if len(schema.EnvConfig) == 0 {
		t.Fatal("Expected environment configuration to be extracted, but got none")
	}

	expectedEnvVars := map[string]struct {
		description string
		example     string
	}{
		"APP_NAME":     {"Application name displayed in UI", "\"Test Frontend\""},
		"PORT":         {"Port for development server", "3000"},
		"API_BASE_URL": {"API base URL", "http://localhost:8000/api"},
	}

	if len(schema.EnvConfig) != len(expectedEnvVars) {
		t.Errorf("Expected %d environment variables, got %d", len(expectedEnvVars), len(schema.EnvConfig))
	}

	// Check each environment variable
	for _, envVar := range schema.EnvConfig {
		if expected, exists := expectedEnvVars[envVar.Name]; exists {
			if envVar.Description != expected.description {
				t.Errorf("EnvVar %s description = '%s', expected '%s'", envVar.Name, envVar.Description, expected.description)
			}
			if envVar.Example != expected.example {
				t.Errorf("EnvVar %s example = '%s', expected '%s'", envVar.Name, envVar.Example, expected.example)
			}
		} else {
			t.Errorf("Unexpected environment variable: %s", envVar.Name)
		}
	}
}

func TestGoAPITemplateExtractWithEnvExample(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "go-api-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal Go API project structure with .env.example
	projectFiles := map[string]string{
		"go.mod": `module github.com/test/api-template

go 1.21`,
		"cmd/api/main.go": `package main

func main() {
	println("Hello, API!")
}`,
		".env.example": `# Server configuration
HTTP_ADDR=:8080
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
# Security
JWT_SECRET=test-secret`,
	}

	// Create the files
	for path, content := range projectFiles {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Test extraction
	goAPI := &GoAPITemplate{}
	schema, err := goAPI.Extract(tempDir)
	if err != nil {
		t.Fatalf("Failed to extract Go API template: %v", err)
	}

	// Verify basic schema properties
	if schema.Type != "go-api" {
		t.Errorf("Expected schema type 'go-api', got '%s'", schema.Type)
	}

	// Verify environment configuration was extracted
	if len(schema.EnvConfig) == 0 {
		t.Fatal("Expected environment configuration to be extracted, but got none")
	}

	expectedEnvVars := []string{"HTTP_ADDR", "DB_HOST", "DB_PORT", "DB_USER", "JWT_SECRET"}
	if len(schema.EnvConfig) != len(expectedEnvVars) {
		t.Errorf("Expected %d environment variables, got %d", len(expectedEnvVars), len(schema.EnvConfig))
	}

	// Check that all expected variables are present
	foundVars := make(map[string]bool)
	for _, envVar := range schema.EnvConfig {
		foundVars[envVar.Name] = true
	}

	for _, expectedVar := range expectedEnvVars {
		if !foundVars[expectedVar] {
			t.Errorf("Expected environment variable %s not found", expectedVar)
		}
	}
}

func TestTemplateExtractWithoutEnvExample(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "no-env-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal project without .env.example
	projectFiles := map[string]string{
		"package.json": `{
  "name": "test-frontend",
  "version": "1.0.0"
}`,
		"src/App.tsx": `export default function App() {
  return <div>Hello World</div>;
}`,
	}

	// Create the files
	for path, content := range projectFiles {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Test extraction
	frontend := &FrontendTemplate{}
	schema, err := frontend.Extract(tempDir)
	if err != nil {
		t.Fatalf("Failed to extract frontend template: %v", err)
	}

	// Verify that empty environment configuration is handled correctly
	if schema.EnvConfig == nil {
		t.Error("Expected EnvConfig to be initialized (empty slice), but got nil")
	}

	if len(schema.EnvConfig) != 0 {
		t.Errorf("Expected no environment variables, got %d", len(schema.EnvConfig))
	}
}
