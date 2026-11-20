package templates

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/acheevo/template-engine/internal/core"
	"github.com/acheevo/template-engine/internal/envparser"
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
	err := filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || path == sourceDir {
			return walkErr
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if r.ShouldSkip(rel) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content, compressed, err := core.CompressContent(string(raw))
		if err != nil {
			return err
		}
		mappings := r.GetMappings(rel)
		schema.Files = append(schema.Files, core.FileSpec{
			Path: rel, Content: content, Compressed: compressed, Template: len(mappings) > 0,
			Size: int64(len(raw)), Hash: core.CalculateContentHash(string(raw)), Mappings: mappings,
		})
		if rel == ".env.example" {
			schema.EnvConfig = envparser.ParseEnvExample(string(raw))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := core.ValidateSchema(schema); err != nil {
		return nil, err
	}
	var digest strings.Builder
	digest.WriteString(schema.Name + schema.Type + schema.Version)
	for _, file := range schema.Files {
		digest.WriteString(file.Path + file.Hash)
	}
	schema.Hash = core.CalculateContentHash(digest.String())
	return schema, nil
}

func (r *ReactNativeTemplate) GetMappings(path string) []core.Mapping {
	switch path {
	case "package.json", "package-lock.json":
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
	if strings.HasPrefix(name, ".env") && name != ".env.example" {
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
	case ".github", ".claude", ".rnstorybook", ".vscode", "assets", "docs", "scripts", "src", "tests":
		return false
	case "AGENTS.md", "LICENSE", ReadmeFile, "app.json", "package.json", "package-lock.json", "eas.json",
		"Dockerfile", "nginx.conf", ".dockerignore", ".gitignore", ".nvmrc", ".env.example", "tsconfig.json",
		"eslint.config.mjs", "jest.config.cjs", "metro.config.js", "playwright.config.ts":
		return false
	default:
		return true
	}
}
