package sdk

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/acheevo/template-engine/internal/core"
)

// FrontendTemplate implements TemplateType for React/frontend projects
type FrontendTemplate struct{}

// Name returns the template type name
func (f *FrontendTemplate) Name() string {
	return "frontend"
}

// Extract analyzes a frontend project and creates a template schema
func (f *FrontendTemplate) Extract(sourceDir string) (*core.TemplateSchema, error) {
	schema := &core.TemplateSchema{
		Name:        "frontend-react-template",
		Type:        "frontend",
		Version:     "1.0.0",
		Description: "React TypeScript frontend template with Tailwind CSS",
		Variables:   f.GetVariables(),
		Files:       []core.FileSpec{},
		Hooks: map[string][]string{
			"post_generate": {"npm install"},
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

		// Process content (compression if needed)
		contentStr := string(content)
		compressedContent, isCompressed, err := core.CompressContent(contentStr)
		if err != nil {
			return err
		}

		// Calculate hash of original content
		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:])

		// Determine if this file needs templating
		isTemplate := f.ShouldTemplate(relPath)

		fileSpec := core.FileSpec{
			Path:       relPath,
			Template:   isTemplate,
			Content:    compressedContent,
			Size:       info.Size(),
			Hash:       hashStr,
			Compressed: isCompressed,
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

	// Calculate schema hash
	schema.Hash = f.calculateSchemaHash(schema)

	return schema, nil
}

// GetMappings returns the variable mappings for a given file path
func (f *FrontendTemplate) GetMappings(filePath string) []core.Mapping {
	return []core.Mapping{
		{Find: "{{.ProjectName}}", Replace: "ProjectName"},
		{Find: "{{.GitHubRepo}}", Replace: "GitHubRepo"},
		{Find: "{{.Author}}", Replace: "Author"},
		{Find: "{{.Description}}", Replace: "Description"},
	}
}

// GetVariables returns the variables used by this template type
func (f *FrontendTemplate) GetVariables() map[string]core.Variable {
	return map[string]core.Variable{
		"ProjectName": {
			Type:        "string",
			Required:    true,
			Description: "Name of the project",
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
			Default:     "A React TypeScript application",
			Description: "Project description",
		},
	}
}

// ShouldTemplate determines if a file should be processed as a template
func (f *FrontendTemplate) ShouldTemplate(filePath string) bool {
	templateFiles := []string{
		"package.json",
		"README.md",
		".env",
		".env.example",
		"index.html",
	}

	fileName := filepath.Base(filePath)
	for _, templateFile := range templateFiles {
		if fileName == templateFile {
			return true
		}
	}

	// Template files in src/ directory
	if strings.Contains(filePath, "src/") && (strings.HasSuffix(filePath, ".js") ||
		strings.HasSuffix(filePath, ".ts") || strings.HasSuffix(filePath, ".jsx") ||
		strings.HasSuffix(filePath, ".tsx")) {
		return true
	}

	return false
}

// ShouldSkip determines if a file should be skipped during extraction
func (f *FrontendTemplate) ShouldSkip(filePath string) bool {
	skipPatterns := []string{
		"node_modules",
		".git",
		".next",
		"build",
		"dist",
		".DS_Store",
		"coverage",
		".nyc_output",
		".cache",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}

	return false
}

// calculateSchemaHash calculates a hash for the entire schema
func (f *FrontendTemplate) calculateSchemaHash(schema *core.TemplateSchema) string {
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

		// Read file content
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
			Content:  string(content),
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

// GetMappings returns the variable mappings for a given file path
func (g *GoAPITemplate) GetMappings(filePath string) []core.Mapping {
	return []core.Mapping{
		{Find: "{{.ProjectName}}", Replace: "ProjectName"},
		{Find: "{{.GitHubRepo}}", Replace: "GitHubRepo"},
		{Find: "{{.Author}}", Replace: "Author"},
		{Find: "{{.Description}}", Replace: "Description"},
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

// ShouldTemplate determines if a file should be processed as a template
func (g *GoAPITemplate) ShouldTemplate(filePath string) bool {
	templateFiles := []string{
		"go.mod",
		"README.md",
		"Dockerfile",
		"docker-compose.yml",
		".env",
		".env.example",
		"Makefile",
	}

	fileName := filepath.Base(filePath)
	for _, templateFile := range templateFiles {
		if fileName == templateFile {
			return true
		}
	}

	// Template Go source files
	if strings.HasSuffix(filePath, ".go") {
		return true
	}

	return false
}

// ShouldSkip determines if a file should be skipped during extraction
func (g *GoAPITemplate) ShouldSkip(filePath string) bool {
	skipPatterns := []string{
		".git",
		"vendor",
		"bin",
		"tmp",
		".DS_Store",
		"coverage",
		"*.log",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}

	return false
}

// calculateSchemaHash calculates a hash for the entire schema
func (g *GoAPITemplate) calculateSchemaHash(schema *core.TemplateSchema) string {
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