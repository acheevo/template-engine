package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/acheevo/template-engine/internal/config"
)

func setupTempConfig(t *testing.T) func() {
	tempDir, err := os.MkdirTemp("", "test-config-*")
	if err != nil {
		t.Fatal(err)
	}

	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	return func() {
		os.Setenv("XDG_CONFIG_HOME", originalXDG)
		os.RemoveAll(tempDir)
	}
}

func TestRunConfigList(t *testing.T) {
	cleanup := setupTempConfig(t)
	defer cleanup()

	err := runConfigList()
	if err != nil {
		t.Errorf("runConfigList() error = %v", err)
	}
}

func TestRunConfigAdd(t *testing.T) {
	cleanup := setupTempConfig(t)
	defer cleanup()

	err := runConfigAdd("test-template", "/test/path", "Test description")
	if err != nil {
		t.Errorf("runConfigAdd() error = %v", err)
	}

	// Verify it was added
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	ref, exists := cfg.References["test-template"]
	if !exists {
		t.Error("Expected test-template to be added")
	}

	if ref.Path != "/test/path" {
		t.Errorf("Expected path '/test/path', got %q", ref.Path)
	}
	if ref.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got %q", ref.Description)
	}
}

func TestRunConfigRemove(t *testing.T) {
	cleanup := setupTempConfig(t)
	defer cleanup()

	// First add a template
	err := runConfigAdd("test-remove", "/test/remove", "Test remove")
	if err != nil {
		t.Fatal(err)
	}

	// Then remove it
	err = runConfigRemove("test-remove")
	if err != nil {
		t.Errorf("runConfigRemove() error = %v", err)
	}

	// Verify it was removed
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if _, exists := cfg.References["test-remove"]; exists {
		t.Error("Expected test-remove to be removed")
	}
}

func TestRunConfigRemoveNonExistent(t *testing.T) {
	cleanup := setupTempConfig(t)
	defer cleanup()

	err := runConfigRemove("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent template")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

func TestRunList(t *testing.T) {
	// This test requires template registration which happens in main
	// Just test that the function doesn't panic
	err := runList()
	if err != nil {
		t.Errorf("runList() error = %v", err)
	}
}

func TestGetReferenceProjectPath(t *testing.T) {
	cleanup := setupTempConfig(t)
	defer cleanup()

	// Create test config
	cfg := &config.ReferenceConfig{
		References: map[string]config.ReferenceProject{
			"test-template": {
				Path:        "/test/path",
				Description: "Test template",
			},
		},
	}

	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	// Test valid template type
	loadedCfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	path, err := loadedCfg.GetReferencePath("test-template")
	if err != nil {
		t.Errorf("GetReferencePath() error = %v", err)
	}

	if path != "/test/path" {
		t.Errorf("Expected path '/test/path', got %q", path)
	}

	// Test invalid template type
	_, err = loadedCfg.GetReferencePath("invalid-type")
	if err == nil {
		t.Error("Expected error for invalid template type")
	}
}
