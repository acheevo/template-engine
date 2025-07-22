package sdk

import (
	"github.com/acheevo/template-engine/internal/core"
)

// createMockClient creates a client with mock templates for testing
func createMockClient() *Client {
	mockFrontendSchema := &core.TemplateSchema{
		Name:        "mock-frontend",
		Type:        "frontend",
		Version:     "1.0.0",
		Description: "Mock frontend template for testing",
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
		},
	}

	mockAPISchema := &core.TemplateSchema{
		Name:        "mock-api",
		Type:        "go-api",
		Version:     "1.0.0",
		Description: "Mock API template for testing",
		Variables: map[string]core.Variable{
			"ProjectName": {Type: "string", Required: true},
			"GitHubRepo":  {Type: "string", Required: true},
		},
		Files: []core.FileSpec{
			{
				Path:     "main.go",
				Template: true,
				Content:  "package main\n\n// {{.ProjectName}}",
				Size:     30,
			},
		},
	}

	return &Client{
		templates: map[string]*core.TemplateSchema{
			"mock-frontend": mockFrontendSchema,
			"mock-api":      mockAPISchema,
		},
	}
}
