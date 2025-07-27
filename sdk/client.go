package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/acheevo/template-engine/internal/core"
	"github.com/acheevo/template-engine/internal/generate"
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

// Generate creates a new project from a registered template
func (c *Client) Generate(ctx context.Context, opts GenerateOptions) error {
	if err := c.ValidateGenerateOptions(opts); err != nil {
		return err
	}

	// Get template schema
	schema, exists := c.templates[opts.Template]
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
func (c *Client) Extract(ctx context.Context, opts ExtractOptions) (*core.TemplateSchema, error) {
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
func (c *Client) GenerateFromTemplate(ctx context.Context, schema *core.TemplateSchema, variables Variables) error {
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
func (c *Client) Validate(schema *core.TemplateSchema) error {
	return core.ValidateSchema(schema)
}

// RegisterTemplate registers a template from a JSON file
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

	// Register the template using its name
	c.templates[schema.Name] = &schema

	return nil
}

// ListTemplates returns available template names and descriptions
func (c *Client) ListTemplates() map[string]string {
	templates := make(map[string]string)
	for name, schema := range c.templates {
		templates[name] = schema.Description
	}
	return templates
}

// GetTemplateInfo returns the structure and metadata for a specific template
func (c *Client) GetTemplateInfo(templateName string) (*TemplateInfo, error) {
	schema, exists := c.templates[templateName]
	if !exists {
		return nil, newTemplateTypeError("GetTemplateInfo", templateName)
	}

	// Convert core.Variable map to SDK Variable map
	sdkVariables := make(map[string]Variable)
	for name, coreVar := range schema.Variables {
		sdkVariables[name] = Variable{
			Type:        coreVar.Type,
			Required:    coreVar.Required,
			Default:     coreVar.Default,
			Description: coreVar.Description,
		}
	}

	return &TemplateInfo{
		Name:        schema.Name,
		Type:        schema.Type,
		Description: schema.Description,
		Variables:   sdkVariables,
	}, nil
}

// GetTemplateVariables returns just the variables for a specific template
func (c *Client) GetTemplateVariables(templateName string) (map[string]Variable, error) {
	schema, exists := c.templates[templateName]
	if !exists {
		return nil, newTemplateTypeError("GetTemplateVariables", templateName)
	}

	// Convert core.Variable map to SDK Variable map
	sdkVariables := make(map[string]Variable)
	for name, coreVar := range schema.Variables {
		sdkVariables[name] = Variable{
			Type:        coreVar.Type,
			Required:    coreVar.Required,
			Default:     coreVar.Default,
			Description: coreVar.Description,
		}
	}

	return sdkVariables, nil
}

// GetTemplateEnvConfig returns the environment configuration for a specific template
func (c *Client) GetTemplateEnvConfig(templateName string) ([]core.EnvVariable, error) {
	schema, exists := c.templates[templateName]
	if !exists {
		return nil, newTemplateTypeError("GetTemplateEnvConfig", templateName)
	}

	return schema.EnvConfig, nil
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

// Variable represents a template variable definition (exported from core.Variable)
type Variable struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
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
