package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/acheevo/template-engine/internal/core"
)

// Generator handles the generation of projects from template schemas
type Generator struct {
	schema          *core.TemplateSchema
	variables       *core.TemplateVariables
	outputDir       string
	templateFuncMap template.FuncMap
}

// NewGenerator creates a new generator instance
func NewGenerator(schemaFile, outputDir, projectName, githubRepo string) (*Generator, error) {
	// Read and parse schema file
	data, err := os.ReadFile(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	var schema core.TemplateSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema file: %w", err)
	}

	// Create template variables
	variables := &core.TemplateVariables{
		ProjectName: projectName,
		GitHubRepo:  githubRepo,
		Author:      "Developer", // Default value
		Description: fmt.Sprintf("A %s application", projectName),
	}

	// Create template function map
	funcMap := template.FuncMap{
		"kebab": func(s string) string {
			return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
		},
		"snake": func(s string) string {
			return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": func(s string) string {
			if s == "" {
				return s
			}
			runes := []rune(s)
			runes[0] = unicode.ToUpper(runes[0])
			return string(runes)
		},
	}

	return &Generator{
		schema:          &schema,
		variables:       variables,
		outputDir:       outputDir,
		templateFuncMap: funcMap,
	}, nil
}

// Generate creates the project from the template schema
func (g *Generator) Generate() error {
	// Validate schema
	if err := core.ValidateSchema(g.schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Validate variables
	if err := core.ValidateVariables(g.schema, g.variables); err != nil {
		return fmt.Errorf("invalid variables: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Process each file in the schema
	for _, fileSpec := range g.schema.Files {
		if err := g.processFile(fileSpec); err != nil {
			return fmt.Errorf("failed to process file %s: %w", fileSpec.Path, err)
		}
	}

	return nil
}

// processFile processes a single file from the schema
func (g *Generator) processFile(fileSpec core.FileSpec) error {
	destPath := filepath.Join(g.outputDir, fileSpec.Path)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	if fileSpec.Template {
		// Process templated file
		return g.processTemplatedFile(fileSpec, destPath)
	} else {
		// Copy static file
		return g.copyStaticFile(fileSpec, destPath)
	}
}

// processTemplatedFile processes a file that needs template substitution
func (g *Generator) processTemplatedFile(fileSpec core.FileSpec, destPath string) error {
	// Decompress content if needed
	content, err := core.DecompressContent(fileSpec.Content, fileSpec.Compressed)
	if err != nil {
		return fmt.Errorf("failed to decompress content: %w", err)
	}

	// Apply mappings first
	for _, mapping := range fileSpec.Mappings {
		content = strings.ReplaceAll(content, mapping.Find, mapping.Replace)
	}

	// Temporarily replace our project template variables and functions with placeholders
	templateReplacements := map[string]string{
		"{{.ProjectName}}":         "__PROJECT_NAME_PLACEHOLDER__",
		"{{.GitHubRepo}}":          "__GITHUB_REPO_PLACEHOLDER__",
		"{{.Author}}":              "__AUTHOR_PLACEHOLDER__",
		"{{.Description}}":         "__DESCRIPTION_PLACEHOLDER__",
		"{{.ProjectName | kebab}}": "__PROJECT_NAME_KEBAB_PLACEHOLDER__",
		"{{.ProjectName | snake}}": "__PROJECT_NAME_SNAKE_PLACEHOLDER__",
		"{{.ProjectName | upper}}": "__PROJECT_NAME_UPPER_PLACEHOLDER__",
		"{{.ProjectName | lower}}": "__PROJECT_NAME_LOWER_PLACEHOLDER__",
		"{{.ProjectName | title}}": "__PROJECT_NAME_TITLE_PLACEHOLDER__",
	}

	for find, replace := range templateReplacements {
		content = strings.ReplaceAll(content, find, replace)
	}

	// Escape all remaining Go template syntax
	content = strings.ReplaceAll(content, "{{", "__ESCAPED_LEFT_BRACE__")
	content = strings.ReplaceAll(content, "}}", "__ESCAPED_RIGHT_BRACE__")

	// Restore our project template variables
	for find, replace := range templateReplacements {
		content = strings.ReplaceAll(content, replace, find)
	}

	// Parse and execute template
	tmpl, err := template.New("file").Funcs(g.templateFuncMap).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template to buffer first
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g.variables); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Restore escaped Go template syntax
	result := buf.String()
	result = strings.ReplaceAll(result, "__ESCAPED_LEFT_BRACE__", "{{")
	result = strings.ReplaceAll(result, "__ESCAPED_RIGHT_BRACE__", "}}")

	// Create destination file and write the final content
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(result); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// copyStaticFile copies a static file that doesn't need templating
func (g *Generator) copyStaticFile(fileSpec core.FileSpec, destPath string) error {
	// Decompress content if needed
	content, err := core.DecompressContent(fileSpec.Content, fileSpec.Compressed)
	if err != nil {
		return fmt.Errorf("failed to decompress content: %w", err)
	}

	// With go-fsck pattern, all content is embedded in the schema
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the embedded content directly
	_, err = file.WriteString(content)
	return err
}

// PrintSummary prints a summary of what was generated
func (g *Generator) PrintSummary() {
	fmt.Printf("Project generated successfully!\n")
	fmt.Printf("Location: %s\n", g.outputDir)
	fmt.Printf("Project Name: %s\n", g.variables.ProjectName)
	fmt.Printf("GitHub Repo: %s\n", g.variables.GitHubRepo)
	fmt.Printf("Files processed: %d\n", len(g.schema.Files))

	templatedCount := 0
	for _, file := range g.schema.Files {
		if file.Template {
			templatedCount++
		}
	}
	fmt.Printf("Templated files: %d\n", templatedCount)
}
