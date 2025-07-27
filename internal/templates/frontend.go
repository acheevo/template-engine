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
		EnvConfig:   []core.EnvVariable{}, // Initialize as empty slice
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

		// Read file content (go-fsck pattern: always include full content)
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
			Content:    compressedContent, // May be compressed
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
func (f *FrontendTemplate) GetMappings(filePath string) []core.Mapping {
	switch filePath {
	case "package.json":
		return []core.Mapping{
			{Find: "\"frontend-template\"", Replace: "\"{{.ProjectName}}\""},
			{Find: "\"Your Name\"", Replace: "\"{{.Author}}\""},
		}
	case "src/config/app.ts":
		return []core.Mapping{
			{Find: "'Frontend Template'", Replace: "'{{.ProjectName}}'"},
			{Find: "'Your Name'", Replace: "'{{.Author}}'"},
		}
	case "README.md":
		return []core.Mapping{
			{Find: "# Frontend Template", Replace: "# {{.ProjectName}}"},
			{Find: "https://github.com/your-username/frontend-template", Replace: "https://github.com/{{.GitHubRepo}}"},
		}
	case "index.html":
		return []core.Mapping{
			{Find: "<title>Frontend Template</title>", Replace: "<title>{{.ProjectName}}</title>"},
		}
	default:
		return []core.Mapping{}
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

// ShouldTemplate determines if a file needs template processing
func (f *FrontendTemplate) ShouldTemplate(filePath string) bool {
	templatedFiles := []string{
		"package.json",
		"README.md",
		"src/config/app.ts",
		"index.html",
	}

	for _, file := range templatedFiles {
		if filePath == file {
			return true
		}
	}

	return false
}

// ShouldSkip determines if a file/directory should be skipped during extraction
func (f *FrontendTemplate) ShouldSkip(path string) bool {
	skipDirs := []string{
		"node_modules",
		"dist",
		"build",
		"coverage",
	}
	return shouldSkipCommon(path, skipDirs)
}

// calculateSchemaHash calculates a hash for the entire schema
func (f *FrontendTemplate) calculateSchemaHash(schema *core.TemplateSchema) string {
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
