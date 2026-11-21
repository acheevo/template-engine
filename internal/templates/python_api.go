package templates

import (
	"path/filepath"
	"strings"

	"github.com/acheevo/template-engine/internal/core"
)

const pythonStarterName = "python-api-template"

// PythonAPITemplate packages the FastAPI and PostgreSQL starter.
type PythonAPITemplate struct{}

func (p *PythonAPITemplate) Name() string { return "python-api" }

func (p *PythonAPITemplate) Extract(sourceDir string) (*core.TemplateSchema, error) {
	schema := &core.TemplateSchema{
		Name:        pythonStarterName,
		Type:        p.Name(),
		Version:     "1.0.0",
		Description: "Python REST API template with FastAPI and PostgreSQL",
		Variables:   p.GetVariables(),
	}
	return extractSourceTree(sourceDir, schema, func(path string, dir bool) bool {
		if dir {
			return p.shouldSkipDirectory(path)
		}
		return p.ShouldSkip(path)
	}, p.GetMappings)
}

func (p *PythonAPITemplate) GetVariables() map[string]core.Variable {
	return map[string]core.Variable{
		"ProjectName": {Type: "string", Required: true, Description: "Name of the API project"},
		"GitHubRepo":  {Type: "string", Required: true, Description: "Destination GitHub repository (owner/name)"},
	}
}

func (p *PythonAPITemplate) GetMappings(path string) []core.Mapping {
	switch path {
	case "pyproject.toml", "uv.lock":
		return []core.Mapping{{
			Find: `name = "` + pythonStarterName + `"`, Replace: `name = {{.ProjectName | slugJSON}}`,
		}}
	case "app/config.py":
		return []core.Mapping{
			{
				Find:    `service_name: str = "` + pythonStarterName + `"`,
				Replace: `service_name: str = {{.ProjectName | slugJSON}}`,
			},
		}
	case envExampleFile:
		return []core.Mapping{{
			Find: "SERVICE_NAME=" + pythonStarterName, Replace: "SERVICE_NAME={{.ProjectName | slugJSON}}",
		}}
	case ReadmeFile:
		return []core.Mapping{
			{Find: "# Python API Template", Replace: "# {{.ProjectName}}"},
			{Find: "github.com/acheevo/" + pythonStarterName, Replace: "github.com/{{.GitHubRepo}}"},
		}
	default:
		return nil
	}
}

func (p *PythonAPITemplate) ShouldTemplate(path string) bool { return len(p.GetMappings(path)) > 0 }

func (p *PythonAPITemplate) ShouldSkip(path string) bool {
	switch path {
	case ReadmeFile, envExampleFile, "AGENTS.md", "LICENSE", "pyproject.toml", "uv.lock", makeFile,
		"Dockerfile", composeFile, ".dockerignore", ".gitignore", ".python-version",
		".github/workflows/ci.yml", ".github/workflows/docker-release.yml",
		".github/dependabot.yml", ".github/ISSUE_TEMPLATE/bug_report.md",
		".github/ISSUE_TEMPLATE/feature_request.md", ".github/PULL_REQUEST_TEMPLATE.md", "docs/openapi.json":
		return false
	}
	root, _, _ := strings.Cut(path, "/")
	return (root != "app" && root != testsDirectory) || filepath.Ext(path) != ".py" ||
		p.shouldSkipDirectory(filepath.ToSlash(filepath.Dir(path))) ||
		strings.HasPrefix(filepath.Base(path), ".")
}

func (*PythonAPITemplate) shouldSkipDirectory(path string) bool {
	switch path {
	case githubDirectory, ".github/workflows", ".github/ISSUE_TEMPLATE", "docs":
		return false
	}
	root, _, _ := strings.Cut(path, "/")
	if root != "app" && root != testsDirectory {
		return true
	}
	for _, part := range strings.Split(path, "/") {
		if strings.HasPrefix(part, ".") || part == "__pycache__" {
			return true
		}
	}
	return false
}
