package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultReferenceConfig(t *testing.T) {
	config := DefaultReferenceConfig()

	if config == nil {
		t.Fatal("DefaultReferenceConfig() returned nil")
	}

	if config.References == nil {
		t.Fatal("References map is nil")
	}

	expectedTypes := []string{"frontend", "go-api"}
	for _, expectedType := range expectedTypes {
		if _, exists := config.References[expectedType]; !exists {
			t.Errorf("Expected template type %q to be in default config", expectedType)
		}
	}

	// Check that default config has reasonable values
	for _, ref := range config.References {
		if ref.Path == "" {
			t.Error("Expected non-empty path in default config")
		}
		if ref.Description == "" {
			t.Error("Expected non-empty description in default config")
		}
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	// Use a temporary directory and set XDG_CONFIG_HOME
	tempDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set environment variable to control config path
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	config, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v, expected nil", err)
	}

	if config == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	// Should return default config when file doesn't exist
	if len(config.References) != 2 {
		t.Errorf("Expected 2 default references, got %d", len(config.References))
	}
}

func TestLoadConfig_ExistingFile(t *testing.T) {
	// Use a temporary directory
	tempDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configDir := filepath.Join(tempDir, "template-engine")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	configFile := filepath.Join(configDir, "references.json")

	// Create a test config
	testConfig := &ReferenceConfig{
		References: map[string]ReferenceProject{
			"test-template": {
				Path:        "/test/path",
				Description: "Test template",
				Version:     "1.0.0",
			},
		},
	}

	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(configFile, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Set environment variable to control config path
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	config, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v, expected nil", err)
	}

	if config == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if len(config.References) != 1 {
		t.Errorf("Expected 1 reference, got %d", len(config.References))
	}

	testRef, exists := config.References["test-template"]
	if !exists {
		t.Error("Expected test-template to exist in loaded config")
	}

	if testRef.Path != "/test/path" {
		t.Errorf("Expected path '/test/path', got %q", testRef.Path)
	}

	if testRef.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %q", testRef.Version)
	}
}

func TestSaveConfig(t *testing.T) {
	// Use a temporary directory
	tempDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "template-engine", "references.json")

	// Set environment variable to control config path
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	testConfig := &ReferenceConfig{
		References: map[string]ReferenceProject{
			"save-test": {
				Path:        "/save/test/path",
				Description: "Save test template",
			},
		},
	}

	err = SaveConfig(testConfig)
	if err != nil {
		t.Errorf("SaveConfig() error = %v, expected nil", err)
	}

	// Verify file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify content
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}

	var loadedConfig ReferenceConfig
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		t.Fatal(err)
	}

	if len(loadedConfig.References) != 1 {
		t.Errorf("Expected 1 reference in saved config, got %d", len(loadedConfig.References))
	}

	saveTest, exists := loadedConfig.References["save-test"]
	if !exists {
		t.Error("Expected save-test to exist in saved config")
	}

	if saveTest.Path != "/save/test/path" {
		t.Errorf("Expected path '/save/test/path', got %q", saveTest.Path)
	}
}

func TestGetReferencePath(t *testing.T) {
	config := &ReferenceConfig{
		References: map[string]ReferenceProject{
			"absolute-path": {
				Path:        "/absolute/path/template",
				Description: "Absolute path template",
			},
			"relative-path": {
				Path:        "../relative/path/template",
				Description: "Relative path template",
			},
		},
	}

	tests := []struct {
		name         string
		templateType string
		wantErr      bool
		checkAbs     bool // Whether to check if result is absolute
	}{
		{
			name:         "existing absolute path",
			templateType: "absolute-path",
			wantErr:      false,
			checkAbs:     true,
		},
		{
			name:         "existing relative path",
			templateType: "relative-path",
			wantErr:      false,
			checkAbs:     true,
		},
		{
			name:         "non-existing template type",
			templateType: "non-existing",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := config.GetReferencePath(tt.templateType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReferencePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if path == "" {
					t.Error("GetReferencePath() returned empty path")
				}

				if tt.checkAbs && !filepath.IsAbs(path) {
					t.Errorf("Expected absolute path, got %q", path)
				}
			}
		})
	}
}

func TestListTemplateTypes(t *testing.T) {
	config := &ReferenceConfig{
		References: map[string]ReferenceProject{
			"type1": {Path: "/path1", Description: "Type 1"},
			"type2": {Path: "/path2", Description: "Type 2"},
			"type3": {Path: "/path3", Description: "Type 3"},
		},
	}

	types := config.ListTemplateTypes()

	if len(types) != 3 {
		t.Errorf("Expected 3 template types, got %d", len(types))
	}

	expectedTypes := map[string]bool{"type1": true, "type2": true, "type3": true}
	for _, templateType := range types {
		if !expectedTypes[templateType] {
			t.Errorf("Unexpected template type: %q", templateType)
		}
		delete(expectedTypes, templateType)
	}

	if len(expectedTypes) > 0 {
		t.Errorf("Missing template types: %v", expectedTypes)
	}
}

func TestAddReference(t *testing.T) {
	config := &ReferenceConfig{}

	config.AddReference("new-template", "/new/path", "New template description")

	if config.References == nil {
		t.Fatal("References map is nil after AddReference")
	}

	if len(config.References) != 1 {
		t.Errorf("Expected 1 reference after AddReference, got %d", len(config.References))
	}

	newRef, exists := config.References["new-template"]
	if !exists {
		t.Error("Expected new-template to exist after AddReference")
	}

	if newRef.Path != "/new/path" {
		t.Errorf("Expected path '/new/path', got %q", newRef.Path)
	}

	if newRef.Description != "New template description" {
		t.Errorf("Expected description 'New template description', got %q", newRef.Description)
	}

	// Test overwriting existing reference
	config.AddReference("new-template", "/updated/path", "Updated description")

	if len(config.References) != 1 {
		t.Errorf("Expected 1 reference after overwrite, got %d", len(config.References))
	}

	updatedRef := config.References["new-template"]
	if updatedRef.Path != "/updated/path" {
		t.Errorf("Expected updated path '/updated/path', got %q", updatedRef.Path)
	}
}
