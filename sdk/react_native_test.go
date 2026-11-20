package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReactNativeExtractAndGenerate(t *testing.T) {
	source := t.TempDir()
	icon := []byte{0x89, 'P', 'N', 'G', 0xff, 0x00, 0x01}
	files := map[string][]byte{
		"app.json":             []byte(`{"expo":{"name": "Mobile App","slug": "mobile-app","scheme": "mobileapp"}}`),
		"package-lock.json":    []byte(`{"name": "mobile-app","packages":{"":{"name": "mobile-app"}}}`),
		".rnstorybook/main.ts": []byte("export default {}"),
		".nvmrc":               []byte("26.13.1\n"),
		"assets/icon.png":      icon,
		"src/empty.ts":         {},
		"src/nested/node_modules/.github/not-copied": []byte("noise"),
		"ios/signing.p12": []byte("private"),
		".env.local":      []byte("SECRET=private"),
	}
	for name, content := range files {
		path := filepath.Join(source, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(source, ".env.local"), filepath.Join(source, "assets/linked.env")); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	client := New()
	schema, err := client.Extract(ctx, ExtractOptions{SourceDir: source, Type: "react-native"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "mobile.json")
	if err := os.WriteFile(archive, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := client.RegisterTemplate(archive); err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	if err := client.Generate(ctx, GenerateOptions{
		Template: "mobile-app-template", ProjectName: `Care "Notes"`, GitHubRepo: "example/care-notes", OutputDir: output,
	}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"assets/icon.png", ".rnstorybook/main.ts", ".nvmrc", "src/empty.ts"} {
		content, err := os.ReadFile(filepath.Join(output, name))
		if err != nil || !bytes.Equal(content, files[name]) {
			t.Fatalf("%s did not round trip: %v", name, err)
		}
	}
	for _, name := range []string{".env.local", "ios", "src/nested/node_modules", "assets/linked.env"} {
		if _, err := os.Stat(filepath.Join(output, name)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("%s should be excluded, got %v", name, err)
		}
	}
	data, err = os.ReadFile(filepath.Join(output, "app.json"))
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Expo struct{ Name, Slug, Scheme string }
	}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	if config.Expo.Name != `Care "Notes"` || config.Expo.Slug != "care-notes" || config.Expo.Scheme != "care-notes" {
		t.Fatalf("unexpected generated Expo identity: %+v", config.Expo)
	}
}
