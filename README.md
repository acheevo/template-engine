# Template Engine

A powerful and flexible template engine for generating projects from templates. Supports both CLI and SDK usage with built-in templates for common project types.

## Features

- **CLI Interface**: Simple commands for quick project generation
- **Go SDK**: Programmatic access for CI/CD and automation
- **Built-in Templates**: Pre-configured templates for frontend and Go API projects
- **Custom Templates**: Extract and reuse templates from existing projects
- **Template Management**: Reference project configuration system
- **Interactive Mode**: Guided project creation

## Installation

### Go Install
```bash
go install github.com/acheevo/template-engine@latest
```

### From Source
```bash
git clone https://github.com/acheevo/template-engine.git
cd template-engine
go build -o template-engine .
```

## Quick Start

### CLI Usage

#### Generate Projects with Built-in Templates
```bash
# Frontend project (React + TypeScript + Vite)
template-engine new frontend "My React App" "user/my-app"

# Go API project (Gin + PostgreSQL + Clean Architecture)
template-engine new go-api "My API Service" "user/my-api"

# Interactive mode
template-engine new --interactive
```

#### Work with Custom Templates
```bash
# Extract a template from existing project
template-engine extract ../my-frontend --type frontend -o my-frontend.json

# Generate project from custom template
template-engine generate my-frontend.json --project-name "Custom App" --github-repo "user/custom-app"

# List available template types
template-engine list
```

#### Manage Reference Projects
```bash
# Add a reference project
template-engine config add frontend /path/to/frontend-template "React frontend template"

# List reference projects
template-engine config list

# Remove a reference project
template-engine config remove frontend
```

### SDK Usage (Go)

#### Basic Project Generation
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/acheevo/template-engine/sdk"
)

func main() {
    client := sdk.New()
    
    // Generate a frontend project
    err := client.Generate(context.Background(), sdk.GenerateOptions{
        Template:    "frontend",
        ProjectName: "My React App",
        GitHubRepo:  "user/my-app",
        OutputDir:   "./my-react-app",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Project generated successfully!")
}
```

#### Custom Template Extraction
```go
// Extract a custom template
template, err := client.Extract(ctx, sdk.ExtractOptions{
    SourceDir: "./my-custom-project",
    Type:      "frontend",
})
if err != nil {
    log.Fatal(err)
}

// Use the extracted template
err = client.GenerateFromTemplate(ctx, template, sdk.Variables{
    ProjectName: "New Project",
    GitHubRepo:  "user/new-project", 
    OutputDir:   "./output",
    Custom: map[string]string{
        "Author": "John Doe",
        "License": "MIT",
    },
})
```

#### CI/CD Integration
```go
// GitHub Actions example
func generateTestEnvironment(prNumber int) error {
    client := sdk.New()
    
    return client.Generate(ctx, sdk.GenerateOptions{
        Template:    "go-api",
        ProjectName: fmt.Sprintf("test-env-%d", prNumber),
        GitHubRepo:  "org/test-repo",
        OutputDir:   fmt.Sprintf("./test-envs/pr-%d", prNumber),
        Variables: map[string]string{
            "Environment": "testing",
            "PRNumber":    strconv.Itoa(prNumber),
        },
    })
}
```

## CLI Commands

### Core Commands
- `template-engine new <type> <name> <repo>` - Generate project from built-in template
- `template-engine extract <source> --type <type>` - Extract template from existing project
- `template-engine generate <template.json>` - Generate project from template file
- `template-engine list` - List available template types

### Configuration Commands
- `template-engine config add <name> <path> <description>` - Add reference project
- `template-engine config list` - List reference projects
- `template-engine config remove <name>` - Remove reference project

### Flags
- `--interactive` - Interactive mode for guided project creation
- `--output <dir>` - Specify output directory
- `--project-name <name>` - Override project name
- `--github-repo <repo>` - Override GitHub repository

## SDK Reference

### Client Creation
```go
client := sdk.New()
```

### Methods

#### Generate(ctx, options) error
Generate a project from a built-in template.

#### Extract(ctx, options) (*Template, error)
Extract a template from an existing project.

#### GenerateFromTemplate(ctx, template, variables) error
Generate a project from a template object.

#### GenerateFromFile(ctx, templateFile, variables) error
Generate a project from a template file.

#### Validate(template) error
Validate a template structure.

#### ListTemplates() ([]string, error)
List available template types.

#### GetTemplateInfo(templateType) (*TemplateInfo, error)
Get complete template metadata and variable structure for a specific template type.

#### GetTemplateVariables(templateType) (map[string]Variable, error)
Get just the variables for a specific template type.

### Template Discovery

You can programmatically discover template structures:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/acheevo/template-engine/sdk"
)

func main() {
    client := sdk.New()
    
    // List all available template types
    templates := client.ListTemplates()
    fmt.Printf("Available templates: %v\n", templates)
    
    // Get detailed info for a specific template
    info, err := client.GetTemplateInfo("frontend")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Template: %s\n", info.Name)
    fmt.Printf("Description: %s\n", info.Description)
    fmt.Printf("Variables:\n")
    
    for name, variable := range info.Variables {
        required := ""
        if variable.Required {
            required = " (required)"
        }
        defaultVal := ""
        if variable.Default != "" {
            defaultVal = fmt.Sprintf(" [default: %s]", variable.Default)
        }
        fmt.Printf("  %s: %s%s%s - %s\n", 
            name, variable.Type, required, defaultVal, variable.Description)
    }
    
    // Or just get variables
    variables, err := client.GetTemplateVariables("go-api")
    if err != nil {
        log.Fatal(err)
    }
    
    for name, variable := range variables {
        fmt.Printf("%s: %+v\n", name, variable)
    }
}
```

### Options Structs

```go
type GenerateOptions struct {
    Template    string            // Template name or path
    ProjectName string            // Name of the project
    GitHubRepo  string            // GitHub repository (org/repo)
    OutputDir   string            // Output directory
    Variables   map[string]string // Custom variables
}

type ExtractOptions struct {
    SourceDir string // Source directory to extract from
    Type      string // Template type
}

type Variables struct {
    ProjectName string            // Project name
    GitHubRepo  string            // GitHub repository
    OutputDir   string            // Output directory
    Custom      map[string]string // Custom variables
}

type TemplateInfo struct {
    Name        string              `json:"name"`
    Type        string              `json:"type"`
    Description string              `json:"description"`
    Variables   map[string]Variable `json:"variables"`
}

type Variable struct {
    Type        string `json:"type"`
    Required    bool   `json:"required"`
    Default     string `json:"default,omitempty"`
    Description string `json:"description,omitempty"`
}
```

### Error Handling

The SDK provides structured error types:

```go
type SDKError struct {
    Type      ErrorType
    Operation string
    Message   string
    Details   string
    Cause     error
}

// Error types
const (
    ErrorTypeValidation ErrorType = "validation"
    ErrorTypeFileSystem ErrorType = "filesystem"
    ErrorTypeTemplate   ErrorType = "template"
    ErrorTypeNetwork    ErrorType = "network"
    ErrorTypeUnknown    ErrorType = "unknown"
)
```

## Real-World Examples

### Microservices Bootstrap
```bash
# Create multiple related services
template-engine new go-api "User Service" "company/user-service"
template-engine new go-api "Order Service" "company/order-service"  
template-engine new go-api "Payment Service" "company/payment-service"
template-engine new frontend "Admin Dashboard" "company/admin-dashboard"
```

### Startup MVP
```bash
# Full stack application
template-engine new frontend "MyApp Web" "startup/myapp-web"
template-engine new go-api "MyApp API" "startup/myapp-api"

cd myapp-web && npm install
cd ../myapp-api && go mod tidy

# Ready to start development!
```

### Custom Template Workflow
```bash
# 1. Extract from your existing "golden" project
template-engine extract ../company-frontend-standard --type frontend -o company-frontend.json

# 2. Generate new projects using your standard
template-engine generate company-frontend.json --project-name "New Client Project" --github-repo "company/client-xyz"

# 3. All new projects follow your standards automatically
```

### Development Team Onboarding
```bash
#!/bin/bash
echo "Setting up development environment..."

# Create local development projects
template-engine new frontend "Local Frontend" "dev/local-frontend" 
template-engine new go-api "Local API" "dev/local-api"

echo "Development environment ready!"
echo "Frontend: cd local-frontend && npm install && npm run dev"
echo "API: cd local-api && go mod tidy && make run"
```

## Project Structure

```
github.com/acheevo/template-engine/
   cmd/                    # CLI commands (Cobra)
      root.go            # Root command setup
      extract.go         # Extract command
      generate.go        # Generate command
      list.go            # List command
      new.go             # New command
      config.go          # Config management commands
   sdk/                   # Go SDK package
      client.go          # Main client
      errors.go          # SDK-specific errors
      mock_test.go       # Test utilities
   internal/
      config/            # Configuration management
          discovery.go   # Reference project discovery
   templates/             # Built-in templates (JSON)
   main.go               # CLI entry point
   README.md             # This file
```

## Template Format

Templates are JSON files with this structure:

```json
{
  "name": "frontend",
  "description": "React TypeScript frontend template",
  "version": "1.0.0",
  "variables": {
    "ProjectName": "string",
    "GitHubRepo": "string",
    "Author": "string"
  },
  "files": [
    {
      "path": "package.json",
      "content": "...",
      "template": true
    }
  ],
  "directories": [
    "src",
    "public",
    "tests"
  ]
}
```

## Configuration

The template engine uses XDG Base Directory standards for configuration:

- **Config File**: `$XDG_CONFIG_HOME/template-engine/config.json`
- **Default Location**: `~/.config/template-engine/config.json`

Configuration includes reference projects that can be used as template sources:

```json
{
  "references": {
    "frontend": {
      "path": "/path/to/frontend-template",
      "description": "Standard React frontend template"
    },
    "go-api": {
      "path": "/path/to/api-template", 
      "description": "Go API with Gin and PostgreSQL"
    }
  }
}
```

## Best Practices

### Template Organization
- Keep template JSON files in a `templates/` directory
- Version your templates (e.g., `frontend-v2.json`)
- Document template variables and their purposes

### Team Usage
- Standardize on template types across your organization
- Extract templates from your best-in-class projects
- Include common configurations, linting, and CI/CD setup

### CI/CD Integration
- Use templates for ephemeral test environments
- Generate documentation sites from templates
- Create standardized deployment configurations

### Maintenance
- Regularly update templates with new best practices
- Test template generation in your CI pipeline
- Keep templates DRY (Don't Repeat Yourself)

## Testing

The project includes comprehensive test coverage:

- **SDK Tests**: Complete unit tests for all SDK functionality
- **CLI Tests**: Command handler tests with proper isolation
- **Configuration Tests**: Config management and persistence
- **CI Compatible**: Self-contained tests with no external dependencies

Run tests:
```bash
# All tests
go test ./...

# Specific package
go test ./sdk/...
go test ./cmd/...
go test ./internal/config/...

# With verbose output
go test ./... -v
```

## Troubleshooting

### Common Issues

**Template not found:**
```bash
# Make sure you have the template files
ls -la *.json
# Or extract them first
template-engine extract ../frontend-template --type frontend -o frontend-template.json
```

**Build errors in generated project:**
```bash
# For frontend projects
cd generated-project && npm install
# For Go projects  
cd generated-project && go mod tidy
```

**Import path issues in Go projects:**
```bash
# Check that the go.mod file has correct module name
head go.mod
# Should show: module github.com/your-repo
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- Create an issue for bug reports or feature requests
- Check existing issues for solutions to common problems
- Review the troubleshooting section above

---

Made with ❤️ by the Acheevo team