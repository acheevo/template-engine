package templates

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/acheevo/template-engine/internal/core"
	"github.com/acheevo/template-engine/internal/envparser"
)

// FullstackTemplate implements TemplateType for fullstack projects with Go API and React frontend
type FullstackTemplate struct{}

// Name returns the template type name
func (f *FullstackTemplate) Name() string {
	return "fullstack"
}

// Extract analyzes a fullstack project and creates a template schema
func (f *FullstackTemplate) Extract(sourceDir string) (*core.TemplateSchema, error) {
	schema := &core.TemplateSchema{
		Name:        "fullstack-template",
		Type:        "fullstack",
		Version:     "1.0.0",
		Description: "Fullstack template with Go API backend and React frontend",
		Variables:   f.GetVariables(),
		Files:       []core.FileSpec{},
		EnvConfig:   []core.EnvVariable{},
		Hooks: map[string][]string{
			"post_generate": {"go mod tidy", "cd frontend && npm install"},
		},
	}

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and files that should be skipped
		if info.IsDir() || f.ShouldSkip(path) {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Calculate hash
		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:])

		// Determine if this file needs templating
		isTemplate := f.ShouldTemplate(relPath)

		fileSpec := core.FileSpec{
			Path:     relPath,
			Template: isTemplate,
			Content:  string(content),
			Size:     info.Size(),
			Hash:     hashStr,
		}

		// Add mappings for templated files
		if isTemplate {
			fileSpec.Mappings = f.GetMappings(relPath)
		}

		schema.Files = append(schema.Files, fileSpec)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Parse .env.example if it exists
	envExamplePath := filepath.Join(sourceDir, ".env.example")
	if _, err := os.Stat(envExamplePath); err == nil {
		envContent, err := os.ReadFile(envExamplePath)
		if err == nil {
			schema.EnvConfig = envparser.ParseEnvExample(string(envContent))
		}
	}

	// Calculate schema hash
	schema.Hash = f.calculateSchemaHash(schema)

	return schema, nil
}

// GetMappings returns the string replacement mappings for a specific file
func (f *FullstackTemplate) GetMappings(filePath string) []core.Mapping {
	switch filePath {
	case "go.mod":
		return []core.Mapping{
			{Find: "module github.com/acheevo/fullstack-template", Replace: "module github.com/{{.GitHubRepo}}"},
		}
	case "cmd/api/main.go":
		return []core.Mapping{
			{Find: "\"github.com/acheevo/fullstack-template/", Replace: "\"github.com/{{.GitHubRepo}}/"},
		}
	case "README.md":
		return []core.Mapping{
			{Find: "# Fullstack Template", Replace: "# {{.ProjectName}}"},
			{Find: "# Go + React Fullstack Template", Replace: "# {{.ProjectName}}"},
			{
				Find:    "git clone https://github.com/acheevo/fullstack-template.git",
				Replace: "git clone https://github.com/{{.GitHubRepo}}.git",
			},
			{Find: "cd fullstack-template", Replace: "cd {{.ProjectName | kebab}}"},
		}
	case "docker-compose.yml":
		return []core.Mapping{
			{Find: "fullstack-template", Replace: "{{.ProjectName | kebab}}"},
			{Find: "fullstack_template", Replace: "{{.ProjectName | snake}}"},
		}
	case "internal/shared/config/config.go":
		return []core.Mapping{
			{
				Find:    "ServiceName    string `envconfig:\"SERVICE_NAME\" default:\"fullstack-template\"`",
				Replace: "ServiceName    string `envconfig:\"SERVICE_NAME\" default:\"{{.ProjectName | kebab}}\"`",
			},
			{
				Find:    "DBName            string `envconfig:\"DB_NAME\" default:\"fullstack_template\"`",
				Replace: "DBName            string `envconfig:\"DB_NAME\" default:\"{{.ProjectName | snake}}\"`",
			},
		}
	case "Makefile":
		return []core.Mapping{
			{Find: "docker build -t fullstack-template", Replace: "docker build -t {{.ProjectName | kebab}}"},
			{Find: "docker rmi fullstack-template", Replace: "docker rmi {{.ProjectName | kebab}}"},
		}
	case "frontend/package.json":
		return []core.Mapping{
			{Find: "\"name\": \"fullstack-template\"", Replace: "\"name\": \"{{.ProjectName | kebab}}\""},
			{Find: "\"description\": \"Fullstack template\"", Replace: "\"description\": \"{{.Description}}\""},
		}
	case "frontend/index.html":
		return []core.Mapping{
			{Find: "<title>Fullstack Template</title>", Replace: "<title>{{.ProjectName}}</title>"},
		}
	case "frontend/src/config/app.ts":
		return []core.Mapping{
			{Find: "APP_NAME: 'Fullstack Template'", Replace: "APP_NAME: '{{.ProjectName}}'"},
		}
	default:
		// Apply global replacements for import paths in all Go files
		if strings.HasSuffix(filePath, ".go") {
			return []core.Mapping{
				{Find: "\"github.com/acheevo/fullstack-template/", Replace: "\"github.com/{{.GitHubRepo}}/"},
			}
		}
		return []core.Mapping{}
	}
}

// GetVariables returns the variables used by this template type
func (f *FullstackTemplate) GetVariables() map[string]core.Variable {
	return map[string]core.Variable{
		"ProjectName": {
			Type:        "string",
			Required:    true,
			Description: "Name of the fullstack project",
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
			Default:     "A fullstack application with Go API and React frontend",
			Description: "Project description",
		},
	}
}

// ShouldTemplate determines if a file needs template processing
func (f *FullstackTemplate) ShouldTemplate(filePath string) bool {
	templatedFiles := []string{
		"go.mod",
		"README.md",
		"docker-compose.yml",
		"cmd/api/main.go",
		"internal/shared/config/config.go",
		"Makefile",
		"frontend/package.json",
		"frontend/index.html",
		"frontend/src/config/app.ts",
	}

	for _, file := range templatedFiles {
		if filePath == file {
			return true
		}
	}

	// Specific Go files that need import path replacements (not all Go files)
	if strings.HasSuffix(filePath, ".go") {
		return true
	}

	return false
}

// ShouldSkip determines if a file/directory should be skipped during extraction
func (f *FullstackTemplate) ShouldSkip(path string) bool {
	baseName := filepath.Base(path)

	// Skip node_modules explicitly (most important skip rule)
	if strings.Contains(path, "node_modules") {
		return true
	}

	// Skip compiled binaries and executables
	if baseName == "api" && !strings.HasSuffix(path, ".go") {
		return true
	}

	// Always include important project dotfiles
	importantDotfiles := []string{
		".dockerignore",
		".gitignore",
		".golangci.yml",
		".golangci.yaml",
		".env.example",
	}

	for _, dotfile := range importantDotfiles {
		if baseName == dotfile {
			return false
		}
	}

	// Always include .claude directory and its contents
	if strings.Contains(path, ".claude") {
		return false
	}

	skipDirs := []string{
		"vendor",
		"bin",
		"tmp",
		"coverage",
		"dist",
		"build",
	}
	return shouldSkipCommon(path, skipDirs)
}

// calculateSchemaHash calculates a hash for the entire schema
func (f *FullstackTemplate) calculateSchemaHash(schema *core.TemplateSchema) string {
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
