package templates

import (
	"path/filepath"
	"strings"

	"github.com/acheevo/template-engine/internal/core"
)

const (
	packageJSONFile = "package.json"
	packageLockFile = "package-lock.json"
	envExampleFile  = ".env.example"
)

// ReactNativeTemplate packages the source of an Expo React Native starter.
type ReactNativeTemplate struct{}

func (r *ReactNativeTemplate) Name() string { return "react-native" }

func (r *ReactNativeTemplate) Extract(sourceDir string) (*core.TemplateSchema, error) {
	schema := &core.TemplateSchema{
		Name:        "mobile-app-template",
		Type:        r.Name(),
		Version:     "1.0.0",
		Description: "Expo and React Native for iOS and Android, with a deployable web preview.",
		Variables:   r.GetVariables(),
	}
	return extractSourceTree(sourceDir, schema, func(path string, _ bool) bool {
		return r.ShouldSkip(path)
	}, r.GetMappings)
}

func (r *ReactNativeTemplate) GetMappings(path string) []core.Mapping {
	switch path {
	case packageJSONFile, packageLockFile:
		return []core.Mapping{{Find: `"name": "mobile-app"`, Replace: `"name": {{.ProjectName | slugJSON}}`}}
	case "app.json":
		return []core.Mapping{
			{Find: `"name": "Mobile App"`, Replace: `"name": {{.ProjectName | jsonString}}`},
			{Find: `"slug": "mobile-app"`, Replace: `"slug": {{.ProjectName | slugJSON}}`},
			{Find: `"scheme": "mobileapp"`, Replace: `"scheme": {{.ProjectName | slugJSON}}`},
		}
	case "src/app/_layout.tsx":
		return []core.Mapping{{Find: `title: 'Mobile App'`, Replace: `title: {{.ProjectName | jsonString}}`}}
	case ReadmeFile:
		return []core.Mapping{{Find: "# Mobile App", Replace: "# {{.ProjectName}}"}}
	default:
		return nil
	}
}

func (r *ReactNativeTemplate) GetVariables() map[string]core.Variable {
	return map[string]core.Variable{
		"ProjectName": {Type: "string", Required: true, Description: "Expo display name and project slug"},
		"GitHubRepo":  {Type: "string", Required: true, Description: "Destination GitHub repository (owner/name)"},
	}
}

func (r *ReactNativeTemplate) ShouldTemplate(path string) bool { return len(r.GetMappings(path)) > 0 }

func (r *ReactNativeTemplate) ShouldSkip(path string) bool {
	for _, part := range strings.Split(path, "/") {
		switch part {
		case ".git", "node_modules", "dist", "dist-storybook", ".expo", "coverage", "test-results", "playwright-report":
			return true
		}
	}
	name := filepath.Base(path)
	if strings.HasPrefix(name, ".env") && name != envExampleFile {
		return true
	}
	switch name {
	case ".DS_Store", "settings.local.json", "credentials.json", "local.properties", "storybook.requires.ts":
		return true
	}
	switch filepath.Ext(name) {
	case ".log", ".pem", ".key", ".p8", ".p12", ".jks", ".keystore", ".mobileprovision":
		return true
	}
	root, _, _ := strings.Cut(path, "/")
	switch root {
	case githubDirectory, ".claude", ".rnstorybook", ".vscode", "assets", "docs", "scripts", "src", testsDirectory:
		return false
	case "AGENTS.md", "LICENSE", ReadmeFile, "app.json", packageJSONFile, packageLockFile, "eas.json",
		"Dockerfile", "nginx.conf", ".dockerignore", ".gitignore", ".nvmrc", envExampleFile, "tsconfig.json",
		"eslint.config.mjs", "jest.config.cjs", "metro.config.js", "playwright.config.ts":
		return false
	default:
		return true
	}
}
