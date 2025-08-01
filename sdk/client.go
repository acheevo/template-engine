package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/acheevo/template-engine/internal/core"
	"github.com/acheevo/template-engine/internal/generate"
	_ "github.com/acheevo/template-engine/internal/templates" // Import to register templates
)

// Client provides programmatic access to the template engine
type Client struct {
	templates map[string]*core.TemplateSchema
}

// New creates a new SDK client
func New() *Client {
	templates := make(map[string]*core.TemplateSchema)

	return &Client{
		templates: templates,
	}
}

// GenerateOptions contains options for generating a project
type GenerateOptions struct {
	Template    string            // Template name (e.g., "frontend", "go-api")
	ProjectName string            // Name of the project
	GitHubRepo  string            // GitHub repository (e.g., "user/repo")
	OutputDir   string            // Output directory
	Variables   map[string]string // Additional template variables
}

// ExtractOptions contains options for extracting a template
type ExtractOptions struct {
	SourceDir string // Source directory to extract from
	Type      string // Template type
	OutputDir string // Optional: directory to save template file
}

// Generate creates a new project from a registered template schema
// Note: This method works with pre-registered template schemas, not template types.
// For template types, use ExtractAndGenerate() workflow instead.
func (c *Client) Generate(ctx context.Context, opts GenerateOptions) error {
	if err := c.ValidateGenerateOptions(opts); err != nil {
		return err
	}

	// Get template schema - try by name first, then by type
	schema, exists := c.templates[opts.Template]
	if !exists {
		// Try to find by template type
		for _, s := range c.templates {
			if s.Type == opts.Template {
				schema = s
				exists = true
				break
			}
		}
	}
	if !exists {
		return newTemplateTypeError("Generate", opts.Template)
	}

	// Create variables from options
	variables := Variables{
		ProjectName: opts.ProjectName,
		GitHubRepo:  opts.GitHubRepo,
		OutputDir:   opts.OutputDir,
		Custom:      opts.Variables,
	}

	// Set defaults if not provided
	if variables.Author == "" {
		variables.Author = "Developer"
	}
	if variables.Description == "" {
		variables.Description = fmt.Sprintf("A %s application", opts.ProjectName)
	}

	return c.GenerateFromTemplate(ctx, schema, variables)
}

// Extract creates a template schema from a source directory using the global registry
func (c *Client) Extract(ctx context.Context, opts ExtractOptions) (*TemplateSchema, error) {
	if err := c.ValidateExtractOptions(opts); err != nil {
		return nil, err
	}

	// Use the global template registry for extraction
	templateType, err := core.GetTemplate(opts.Type)
	if err != nil {
		return nil, newTemplateTypeError("Extract", opts.Type)
	}

	schema, err := templateType.Extract(opts.SourceDir)
	if err != nil {
		return nil, newExtractionError("Extract", "failed to extract template from source directory", err)
	}

	return schema, nil
}

// GenerateFromTemplate creates a project from a template schema
func (c *Client) GenerateFromTemplate(ctx context.Context, schema *TemplateSchema, variables Variables) error {
	if err := c.ValidateVariables(variables); err != nil {
		return err
	}

	if err := c.Validate(schema); err != nil {
		return newSchemaError("GenerateFromTemplate", "invalid template schema", err)
	}

	// Create temporary file for the schema
	tempFile, err := os.CreateTemp("", "template-schema-*.json")
	if err != nil {
		return newFileSystemError("GenerateFromTemplate", "failed to create temporary file", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Marshal schema to JSON
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return newSchemaError("GenerateFromTemplate", "failed to marshal schema to JSON", err)
	}

	// Write schema to temporary file
	if _, err := tempFile.Write(schemaJSON); err != nil {
		return newFileSystemError("GenerateFromTemplate", "failed to write schema file", err)
	}
	tempFile.Close()

	// Create generator (reuse existing logic)
	generator, err := generate.NewGenerator(tempFile.Name(), variables.OutputDir,
		variables.ProjectName, variables.GitHubRepo)
	if err != nil {
		return newGenerationError("GenerateFromTemplate", "failed to create generator", err)
	}

	if err := generator.Generate(); err != nil {
		return newGenerationError("GenerateFromTemplate", "failed to generate project", err)
	}

	return nil
}

// Validate checks if a template schema is valid
func (c *Client) Validate(schema *TemplateSchema) error {
	return core.ValidateSchema(schema)
}

// RegisterTemplate registers a template schema from a JSON file for use with Generate()
// This is for working with pre-extracted template schema files, not template types.
// Template types are automatically registered via the global registry.
func (c *Client) RegisterTemplate(templatePath string) error {
	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return newFileSystemError("RegisterTemplate", "template file does not exist", err)
	}

	// Load template schema from file
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return newFileSystemError("RegisterTemplate", "failed to read template file", err)
	}

	var schema core.TemplateSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return newSchemaError("RegisterTemplate", "failed to parse template file", err)
	}

	// Validate the schema
	if err := c.Validate(&schema); err != nil {
		return newSchemaError("RegisterTemplate", "invalid template schema", err)
	}

	// Register the template using its name in the client's local cache
	// This is separate from the global template type registry
	c.templates[schema.Name] = &schema

	return nil
}

// ========================================
// Template Types API (Built-in Extractors)
// ========================================

// ListTemplateTypes returns available built-in template types for extraction
func (c *Client) ListTemplateTypes() []string {
	return core.ListTemplates()
}

// GetTemplateTypeInfo returns metadata for a built-in template type
func (c *Client) GetTemplateTypeInfo(templateType string) (*TemplateTypeInfo, error) {
	tmpl, err := core.GetTemplate(templateType)
	if err != nil {
		return nil, newTemplateTypeError("GetTemplateTypeInfo", templateType)
	}

	return &TemplateTypeInfo{
		Name:        tmpl.Name(),
		Description: fmt.Sprintf("%s template type", tmpl.Name()),
		Variables:   tmpl.GetVariables(), // Direct use since Variable = core.Variable
	}, nil
}

// ExtractSchema extracts a template schema from a source directory using a template type
func (c *Client) ExtractSchema(templateType, sourceDir string) (*TemplateSchema, error) {
	return c.Extract(context.Background(), ExtractOptions{
		SourceDir: sourceDir,
		Type:      templateType,
	})
}

// ExtractAndGenerateFromType is a convenience method that extracts and generates in one step
func (c *Client) ExtractAndGenerateFromType(templateType, sourceDir, projectName, githubRepo, outputDir string) error {
	return c.ExtractAndGenerate(context.Background(), sourceDir, templateType, projectName, githubRepo, outputDir)
}

// ========================================
// Template Schemas API (User-registered Data)
// ========================================

// RegisterSchema registers a template schema from a JSON file
func (c *Client) RegisterSchema(schemaFile string) error {
	return c.RegisterTemplate(schemaFile) // Delegate to existing method
}

// ListSchemas returns registered template schema names
func (c *Client) ListSchemas() []string {
	names := make([]string, 0, len(c.templates))
	for name := range c.templates {
		names = append(names, name)
	}
	return names
}

// GetSchemaInfo returns detailed information about a registered template schema
func (c *Client) GetSchemaInfo(schemaName string) (*TemplateSchemaInfo, error) {
	schema, exists := c.templates[schemaName]
	if !exists {
		return nil, newTemplateTypeError("GetSchemaInfo", schemaName)
	}

	return &TemplateSchemaInfo{
		Name:        schema.Name,
		Type:        schema.Type,
		Version:     schema.Version,
		Description: schema.Description,
		Variables:   schema.Variables, // Direct use since Variable = core.Variable
		FileCount:   len(schema.Files),
		EnvVarCount: len(schema.EnvConfig),
	}, nil
}

// GetSchemaEnvConfig returns environment configuration for a registered template schema
func (c *Client) GetSchemaEnvConfig(schemaName string) ([]EnvVariable, error) {
	schema, exists := c.templates[schemaName]
	if !exists {
		return nil, newTemplateTypeError("GetSchemaEnvConfig", schemaName)
	}
	return schema.EnvConfig, nil
}

// GenerateFromSchema generates a project from a registered template schema
func (c *Client) GenerateFromSchema(ctx context.Context, schemaName string, variables Variables) error {
	schema, exists := c.templates[schemaName]
	if !exists {
		return newTemplateTypeError("GenerateFromSchema", schemaName)
	}
	return c.GenerateFromTemplate(ctx, schema, variables)
}

// Variables contains template variables
type Variables struct {
	ProjectName string
	GitHubRepo  string
	OutputDir   string
	Author      string
	Description string
	Custom      map[string]string
}

// TemplateInfo represents template metadata and structure
type TemplateInfo struct {
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Description string              `json:"description"`
	Variables   map[string]Variable `json:"variables"`
}

// Type aliases to avoid repetitive conversions
type (
	Variable       = core.Variable
	EnvVariable    = core.EnvVariable
	TemplateSchema = core.TemplateSchema
)

// TemplateTypeInfo represents metadata for a built-in template type (extractor)
type TemplateTypeInfo struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Variables   map[string]Variable `json:"variables"`
}

// TemplateSchemaInfo represents detailed information about a registered template schema
type TemplateSchemaInfo struct {
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Variables   map[string]Variable `json:"variables"`
	FileCount   int                 `json:"file_count"`
	EnvVarCount int                 `json:"env_var_count"`
}

// ExtractAndGenerate extracts a template from a source directory and immediately generates a project
// This is the main workflow method that combines extraction and generation in one step
func (c *Client) ExtractAndGenerate(ctx context.Context, sourceDir, templateType,
	projectName, githubRepo, outputDir string,
) error {
	// Validate inputs
	if sourceDir == "" {
		return newValidationError("ExtractAndGenerate", "source directory is required", "")
	}
	if templateType == "" {
		return newValidationError("ExtractAndGenerate", "template type is required", "")
	}
	if projectName == "" {
		return newValidationError("ExtractAndGenerate", "project name is required", "")
	}
	if githubRepo == "" {
		return newValidationError("ExtractAndGenerate", "github repo is required", "")
	}
	if outputDir == "" {
		return newValidationError("ExtractAndGenerate", "output directory is required", "")
	}

	// Check if source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return newFileSystemError("ExtractAndGenerate", "source directory does not exist", err)
	}

	// Step 1: Extract template schema from source directory
	schema, err := c.Extract(ctx, ExtractOptions{
		SourceDir: sourceDir,
		Type:      templateType,
	})
	if err != nil {
		return err // Error already wrapped by Extract method
	}

	// Step 2: Generate project from extracted schema
	variables := Variables{
		ProjectName: projectName,
		GitHubRepo:  githubRepo,
		OutputDir:   outputDir,
		Author:      "Developer", // Default value
		Description: fmt.Sprintf("A %s application", projectName),
	}

	err = c.GenerateFromTemplate(ctx, schema, variables)
	if err != nil {
		return err // Error already wrapped by GenerateFromTemplate method
	}

	return nil
}

// GenerateFromFile loads a template schema from a file and generates a project
// This is a convenience method for when you already have a template.json file
func (c *Client) GenerateFromFile(ctx context.Context, templateFile string, variables Variables) error {
	if err := c.ValidateVariables(variables); err != nil {
		return err
	}

	// Check if template file exists
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		return newFileSystemError("GenerateFromFile", "template file does not exist", err)
	}

	// Load template schema from file
	data, err := os.ReadFile(templateFile)
	if err != nil {
		return newFileSystemError("GenerateFromFile", "failed to read template file", err)
	}

	var schema core.TemplateSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return newSchemaError("GenerateFromFile", "failed to parse template file", err)
	}

	// Generate from the loaded schema
	return c.GenerateFromTemplate(ctx, &schema, variables)
}

// ValidateGenerateOptions validates GenerateOptions
func (c *Client) ValidateGenerateOptions(opts GenerateOptions) error {
	if opts.ProjectName == "" {
		return newValidationError("Generate", "project name is required", "")
	}
	if opts.GitHubRepo == "" {
		return newValidationError("Generate", "github repo is required", "")
	}
	if opts.Template == "" {
		return newValidationError("Generate", "template type is required", "")
	}
	return nil
}

// ValidateExtractOptions validates ExtractOptions
func (c *Client) ValidateExtractOptions(opts ExtractOptions) error {
	if opts.SourceDir == "" {
		return newValidationError("Extract", "source directory is required", "")
	}
	if opts.Type == "" {
		return newValidationError("Extract", "template type is required", "")
	}
	// Check if source directory exists
	if _, err := os.Stat(opts.SourceDir); os.IsNotExist(err) {
		return newFileSystemError("Extract", "source directory does not exist", err)
	}
	return nil
}

// ValidateVariables validates Variables
func (c *Client) ValidateVariables(variables Variables) error {
	if variables.ProjectName == "" {
		return newValidationError("GenerateFromTemplate", "project name is required", "")
	}
	if variables.GitHubRepo == "" {
		return newValidationError("GenerateFromTemplate", "github repo is required", "")
	}
	if variables.OutputDir == "" {
		return newValidationError("GenerateFromTemplate", "output directory is required", "")
	}
	return nil
}
