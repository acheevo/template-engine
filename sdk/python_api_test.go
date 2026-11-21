package sdk

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPythonAPIExtractAndGenerate(t *testing.T) {
	source := t.TempDir()
	const starter = `name = "python-api-template"`
	files := map[string]string{
		"pyproject.toml":           "[project]\n" + starter + "\n",
		"uv.lock":                  "[[package]]\n" + starter + "\n",
		"app/config.py":            `service_name: str = "python-api-template"`,
		"app/nested/__init__.py":   "",
		".env.example":             "SERVICE_NAME=python-api-template\n",
		"README.md":                "# Python API Template\nhttps://github.com/acheevo/python-api-template",
		".python-version":          "3.14\n",
		".github/workflows/ci.yml": "build: ${{ github.ref }}\n",
		"docs/openapi.json":        "{}",
		".env":                     "JWT_SECRET=local-only\n",
		".venv/lib/site.py":        "local-only\n",
		"app/.env.production":      "local-only\n",
		"app/__pycache__/local.py": "local-only\n",
		"app/token.pem":            "local-only\n",
		"docs/cache.json":          "local-only\n",
	}
	for name, content := range files {
		path := filepath.Join(source, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	ctx := context.Background()
	client := New()
	schema, err := client.Extract(ctx, ExtractOptions{SourceDir: source, Type: "python-api"})
	if err != nil {
		t.Fatal(err)
	}
	if schema.Name != "python-api-template" || schema.Type != "python-api" ||
		len(schema.EnvConfig) != 1 || schema.EnvConfig[0].Name != "SERVICE_NAME" {
		t.Fatalf("incorrect template metadata: %s %s %+v", schema.Name, schema.Type, schema.EnvConfig)
	}
	output := t.TempDir()
	if err := client.GenerateFromTemplate(ctx, schema, Variables{
		ProjectName: `23 "Reports" API`, GitHubRepo: "sample/reports", OutputDir: output,
	}); err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"pyproject.toml":           `name = "app-23-reports-api"`,
		"uv.lock":                  `name = "app-23-reports-api"`,
		"app/config.py":            `service_name: str = "app-23-reports-api"`,
		"app/nested/__init__.py":   "",
		".env.example":             `SERVICE_NAME="app-23-reports-api"`,
		"README.md":                "# 23 \"Reports\" API\nhttps://github.com/sample/reports",
		".python-version":          "3.14\n",
		".github/workflows/ci.yml": files[".github/workflows/ci.yml"],
		"docs/openapi.json":        "{}",
	}
	for name, want := range expected {
		data, err := os.ReadFile(filepath.Join(output, name))
		if err != nil || !strings.Contains(string(data), want) {
			t.Errorf("generated %s = %q, %v, want %q", name, data, err, want)
		}
	}
	for name := range files {
		if _, wanted := expected[name]; wanted {
			continue
		}
		if _, err := os.Stat(filepath.Join(output, name)); !os.IsNotExist(err) {
			t.Errorf("local file %s was not excluded: %v", name, err)
		}
	}
}
