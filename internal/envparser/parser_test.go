package envparser

import (
	"testing"

	"github.com/acheevo/template-engine/internal/core"
)

func TestParseEnvExample(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []core.EnvVariable
	}{
		{
			name: "simple env variables",
			content: `# Database host
DB_HOST=localhost
# Database port
DB_PORT=5432`,
			expected: []core.EnvVariable{
				{Name: "DB_HOST", Description: "Database host", Example: "localhost"},
				{Name: "DB_PORT", Description: "Database port", Example: "5432"},
			},
		},
		{
			name: "variables without descriptions",
			content: `APP_NAME=myapp
PORT=3000`,
			expected: []core.EnvVariable{
				{Name: "APP_NAME", Description: "", Example: "myapp"},
				{Name: "PORT", Description: "", Example: "3000"},
			},
		},
		{
			name: "variables with complex descriptions",
			content: `# Secret key for signing JWT tokens (CHANGE IN PRODUCTION!)
JWT_SECRET=your-jwt-secret-key
# Maximum duration for reading the entire request
READ_TIMEOUT=15s`,
			expected: []core.EnvVariable{
				{
					Name:        "JWT_SECRET",
					Description: "Secret key for signing JWT tokens (CHANGE IN PRODUCTION!)",
					Example:     "your-jwt-secret-key",
				},
				{Name: "READ_TIMEOUT", Description: "Maximum duration for reading the entire request", Example: "15s"},
			},
		},
		{
			name: "variables with quoted values",
			content: `# Project name
PROJECT_NAME="My Project"
# API URL
API_URL="http://localhost:8000"`,
			expected: []core.EnvVariable{
				{Name: "PROJECT_NAME", Description: "Project name", Example: "\"My Project\""},
				{Name: "API_URL", Description: "API URL", Example: "\"http://localhost:8000\""},
			},
		},
		{
			name: "empty lines and multiple comments",
			content: `# Database Configuration
# Database host address
DB_HOST=localhost

# Database port number
DB_PORT=5432

# Service Configuration
SERVICE_NAME=myservice`,
			expected: []core.EnvVariable{
				{Name: "DB_HOST", Description: "Database host address", Example: "localhost"},
				{Name: "DB_PORT", Description: "Database port number", Example: "5432"},
				{Name: "SERVICE_NAME", Description: "Service Configuration", Example: "myservice"},
			},
		},
		{
			name: "variable without value",
			content: `# Optional service ID
SERVICE_ID=`,
			expected: []core.EnvVariable{
				{Name: "SERVICE_ID", Description: "Optional service ID", Example: ""},
			},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []core.EnvVariable{},
		},
		{
			name: "comments only",
			content: `# This is just a comment
# Another comment`,
			expected: []core.EnvVariable{},
		},
		{
			name: "malformed lines ignored",
			content: `# Valid variable
VALID_VAR=value
invalid line without equals
# Another valid variable
ANOTHER_VAR=another_value`,
			expected: []core.EnvVariable{
				{Name: "VALID_VAR", Description: "Valid variable", Example: "value"},
				{Name: "ANOTHER_VAR", Description: "Another valid variable", Example: "another_value"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEnvExample(tt.content)

			if len(result) != len(tt.expected) {
				t.Errorf("ParseEnvExample() returned %d variables, expected %d", len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing expected variable at index %d: %+v", i, expected)
					continue
				}

				actual := result[i]
				if actual.Name != expected.Name {
					t.Errorf("Variable %d name = %v, expected %v", i, actual.Name, expected.Name)
				}
				if actual.Description != expected.Description {
					t.Errorf("Variable %d description = %v, expected %v", i, actual.Description, expected.Description)
				}
				if actual.Example != expected.Example {
					t.Errorf("Variable %d example = %v, expected %v", i, actual.Example, expected.Example)
				}
			}
		})
	}
}

func TestParseEnvExampleRealWorld(t *testing.T) {
	// Test with actual .env.example content similar to our templates
	content := `# HTTP Server Configuration
# Server listen address and port
HTTP_ADDR=:8080
# Maximum duration for reading the entire request
READ_TIMEOUT=15s

# Database Configuration
# Database host address
DB_HOST=localhost
# Database port number
DB_PORT=5432
# Database username for authentication
DB_USER=postgres
# Database password for authentication
DB_PASSWORD=postgres

# Security Configuration
# Secret key for signing JWT tokens (CHANGE IN PRODUCTION!)
JWT_SECRET=your-jwt-secret-key
# JWT token expiration duration
JWT_EXPIRATION=24h

# Observability
# Enable Prometheus metrics collection and /metrics endpoint
METRICS_ENABLED=true`

	result := ParseEnvExample(content)

	expectedCount := 9
	if len(result) != expectedCount {
		t.Errorf("Expected %d variables, got %d", expectedCount, len(result))
	}

	// Test a few specific variables
	expectedVars := map[string]struct {
		description string
		example     string
	}{
		"HTTP_ADDR":       {"Server listen address and port", ":8080"},
		"DB_HOST":         {"Database host address", "localhost"},
		"JWT_SECRET":      {"Secret key for signing JWT tokens (CHANGE IN PRODUCTION!)", "your-jwt-secret-key"},
		"METRICS_ENABLED": {"Enable Prometheus metrics collection and /metrics endpoint", "true"},
	}

	foundVars := make(map[string]core.EnvVariable)
	for _, envVar := range result {
		foundVars[envVar.Name] = envVar
	}

	for name, expected := range expectedVars {
		if actual, found := foundVars[name]; found {
			if actual.Description != expected.description {
				t.Errorf("Variable %s description = %v, expected %v", name, actual.Description, expected.description)
			}
			if actual.Example != expected.example {
				t.Errorf("Variable %s example = %v, expected %v", name, actual.Example, expected.example)
			}
		} else {
			t.Errorf("Expected variable %s not found", name)
		}
	}
}
