package core

import (
	"fmt"
)

// TemplateRegistry manages different template types
type TemplateRegistry struct {
	templates map[string]TemplateType
}

// NewTemplateRegistry creates a new template registry
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]TemplateType),
	}
}

// Register adds a template type to the registry
func (r *TemplateRegistry) Register(templateType TemplateType) {
	r.templates[templateType.Name()] = templateType
}

// Get retrieves a template type by name
func (r *TemplateRegistry) Get(name string) (TemplateType, error) {
	template, exists := r.templates[name]
	if !exists {
		return nil, fmt.Errorf("template type not found: %s", name)
	}
	return template, nil
}

// List returns all registered template types
func (r *TemplateRegistry) List() []string {
	names := make([]string, 0, len(r.templates))
	for name := range r.templates {
		names = append(names, name)
	}
	return names
}

// Global registry instance
var GlobalRegistry = NewTemplateRegistry()

// RegisterTemplate registers a template type globally
func RegisterTemplate(templateType TemplateType) {
	GlobalRegistry.Register(templateType)
}

// GetTemplate retrieves a template type from global registry
func GetTemplate(name string) (TemplateType, error) {
	return GlobalRegistry.Get(name)
}

// ListTemplates returns all registered template types
func ListTemplates() []string {
	return GlobalRegistry.List()
}
