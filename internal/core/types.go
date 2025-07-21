package core

// TemplateSchema represents the complete template configuration
type TemplateSchema struct {
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Variables   map[string]Variable `json:"variables"`
	Files       []FileSpec          `json:"files"`
	Hooks       map[string][]string `json:"hooks,omitempty"`
	Hash        string              `json:"hash,omitempty"`
}

// Variable represents a template variable definition
type Variable struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// FileSpec represents a file in the template (go-fsck pattern: all content embedded)
type FileSpec struct {
	Path       string    `json:"path"`
	Template   bool      `json:"template"`
	Content    string    `json:"content"`              // Always includes full content
	Size       int64     `json:"size"`                 // Original file size
	Hash       string    `json:"hash,omitempty"`       // Content hash for validation
	Compressed bool      `json:"compressed,omitempty"` // If content is compressed
	Mappings   []Mapping `json:"mappings,omitempty"`
}

// Mapping represents a string replacement mapping
type Mapping struct {
	Find    string `json:"find"`
	Replace string `json:"replace"`
}

// TemplateVariables represents the variables to substitute during generation
type TemplateVariables struct {
	ProjectName string `json:"project_name"`
	GitHubRepo  string `json:"github_repo"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description,omitempty"`
}

// TemplateType represents different types of templates (frontend, go-api, etc.)
type TemplateType interface {
	Name() string
	Extract(sourceDir string) (*TemplateSchema, error)
	GetMappings(filePath string) []Mapping
	GetVariables() map[string]Variable
	ShouldTemplate(filePath string) bool
	ShouldSkip(filePath string) bool
}
