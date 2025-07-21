package sdk

import (
	"github.com/acheevo/template-engine/internal/core"
)

// mockTemplateType implements core.TemplateType for testing
type mockTemplateType struct {
	name string
}

func (m *mockTemplateType) Name() string {
	return m.name
}

func (m *mockTemplateType) Extract(sourceDir string) (*core.TemplateSchema, error) {
	return &core.TemplateSchema{
		Name:        "mock-template",
		Type:        m.name,
		Version:     "1.0.0",
		Description: "Mock template for testing",
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
	}, nil
}

func (m *mockTemplateType) GetMappings(filePath string) []core.Mapping {
	return nil
}

func (m *mockTemplateType) GetVariables() map[string]core.Variable {
	return map[string]core.Variable{
		"ProjectName": {Type: "string", Required: true},
		"GitHubRepo":  {Type: "string", Required: true},
	}
}

func (m *mockTemplateType) ShouldTemplate(filePath string) bool {
	return true
}

func (m *mockTemplateType) ShouldSkip(filePath string) bool {
	return false
}

// createMockClient creates a client with mock templates for testing
func createMockClient() *Client {
	return &Client{
		templates: map[string]core.TemplateType{
			"mock-frontend": &mockTemplateType{name: "mock-frontend"},
			"mock-api":      &mockTemplateType{name: "mock-api"},
		},
	}
}
