package templates

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/acheevo/template-engine/internal/core"
)

// GoAPITemplate implements TemplateType for Go API projects
type GoAPITemplate struct{}

// Name returns the template type name
func (g *GoAPITemplate) Name() string {
	return "go-api"
}

// Extract analyzes a Go API project and creates a template schema
func (g *GoAPITemplate) Extract(sourceDir string) (*core.TemplateSchema, error) {
	schema := &core.TemplateSchema{
		Name:        "go-api-template",
		Type:        "go-api",
		Version:     "1.0.0",
		Description: "Go REST API template with Gin and PostgreSQL",
		Variables:   g.GetVariables(),
		Files:       []core.FileSpec{},
		Hooks: map[string][]string{
			"post_generate": {"go mod tidy", "go build"},
		},
	}

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and files that should be skipped
		if info.IsDir() || g.ShouldSkip(path) {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Read file content (go-fsck pattern: always include full content)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Calculate hash
		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:])

		// Determine if this file needs templating
		isTemplate := g.ShouldTemplate(relPath)

		fileSpec := core.FileSpec{
			Path:     relPath,
			Template: isTemplate,
			Content:  string(content), // Always include full content
			Size:     info.Size(),
			Hash:     hashStr,
		}

		// Add mappings for templated files
		if isTemplate {
			fileSpec.Mappings = g.GetMappings(relPath)
		}

		schema.Files = append(schema.Files, fileSpec)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Calculate schema hash
	schema.Hash = g.calculateSchemaHash(schema)

	return schema, nil
}

// GetMappings returns the string replacement mappings for a specific file
func (g *GoAPITemplate) GetMappings(filePath string) []core.Mapping {
	switch filePath {
	case "go.mod":
		return []core.Mapping{
			{Find: "module github.com/acheevo/api-template", Replace: "module github.com/{{.GitHubRepo}}"},
		}
	case "cmd/api/main.go":
		return []core.Mapping{
			{Find: "\"github.com/acheevo/api-template/", Replace: "\"github.com/{{.GitHubRepo}}/"},
		}
	case "README.md":
		return []core.Mapping{
			{Find: "# Go API Template", Replace: "# {{.ProjectName}}"},
			{
				Find:    "git clone https://github.com/acheevo/api-template.git",
				Replace: "git clone https://github.com/{{.GitHubRepo}}.git",
			},
			{Find: "cd api-template", Replace: "cd {{.ProjectName | kebab}}"},
		}
	case "docker-compose.yml":
		return []core.Mapping{
			{Find: "api-template", Replace: "{{.ProjectName | kebab}}"},
		}
	case "internal/shared/config/config.go":
		return []core.Mapping{
			{
				Find:    "ServiceName    string `envconfig:\"SERVICE_NAME\" default:\"api-template\"`",
				Replace: "ServiceName    string `envconfig:\"SERVICE_NAME\" default:\"{{.ProjectName | kebab}}\"`",
			},
			{
				Find:    "DBName            string `envconfig:\"DB_NAME\" default:\"api_template\"`",
				Replace: "DBName            string `envconfig:\"DB_NAME\" default:\"{{.ProjectName | lower}}\"`",
			},
		}
	case "Makefile":
		return []core.Mapping{
			{Find: "docker build -t api-template .", Replace: "docker build -t {{.ProjectName | kebab}} ."},
			{Find: "docker rmi api-template", Replace: "docker rmi {{.ProjectName | kebab}}"},
		}
	default:
		// Apply global replacements for import paths in all Go files
		if strings.HasSuffix(filePath, ".go") {
			return []core.Mapping{
				{Find: "\"github.com/acheevo/api-template/", Replace: "\"github.com/{{.GitHubRepo}}/"},
			}
		}
		return []core.Mapping{}
	}
}

// GetVariables returns the variables used by this template type
func (g *GoAPITemplate) GetVariables() map[string]core.Variable {
	return map[string]core.Variable{
		"ProjectName": {
			Type:        "string",
			Required:    true,
			Description: "Name of the API project",
		},
		"GitHubRepo": {
			Type:        "string",
			Required:    true,
			Description: "GitHub repository (e.g., username/repo-name)",
		},
		"Author": {
			Type:        "string",
			Required:    false,
			Default:     "Developer",
			Description: "Project author name",
		},
		"Description": {
			Type:        "string",
			Required:    false,
			Default:     "A Go REST API application",
			Description: "Project description",
		},
	}
}

// ShouldTemplate determines if a file needs template processing
func (g *GoAPITemplate) ShouldTemplate(filePath string) bool {
	templatedFiles := []string{
		"go.mod",
		"README.md",
		"docker-compose.yml",
		"cmd/api/main.go",
		"internal/shared/config/config.go",
		"Makefile",
	}

	for _, file := range templatedFiles {
		if filePath == file {
			return true
		}
	}

	// All Go files need import path replacements
	if strings.HasSuffix(filePath, ".go") {
		return true
	}

	return false
}

// ShouldSkip determines if a file/directory should be skipped during extraction
func (g *GoAPITemplate) ShouldSkip(path string) bool {
	skipPatterns := []string{
		".git",
		".DS_Store",
		"vendor",
		"bin",
		"tmp",
		"*.log",
		".env",
		"coverage.out",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

// calculateSchemaHash calculates a hash for the entire schema
func (g *GoAPITemplate) calculateSchemaHash(schema *core.TemplateSchema) string {
	// Create a deterministic string representation of the schema
	var content strings.Builder
	content.WriteString(schema.Name)
	content.WriteString(schema.Type)
	content.WriteString(schema.Version)

	for _, file := range schema.Files {
		content.WriteString(file.Path)
		content.WriteString(file.Hash)
	}

	hash := sha256.Sum256([]byte(content.String()))
	return hex.EncodeToString(hash[:])
}
