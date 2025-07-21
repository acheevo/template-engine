package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReferenceConfig defines where reference projects are located
type ReferenceConfig struct {
	References map[string]ReferenceProject `json:"references"`
}

// ReferenceProject defines a reference project location and metadata
type ReferenceProject struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	Version     string `json:"version,omitempty"`
}

// DefaultReferenceConfig returns the default configuration
func DefaultReferenceConfig() *ReferenceConfig {
	return &ReferenceConfig{
		References: map[string]ReferenceProject{
			"frontend": {
				Path:        "../frontend-template",
				Description: "React + TypeScript + Vite frontend template",
			},
			"go-api": {
				Path:        "../api-template",
				Description: "Go API with Gin + PostgreSQL + Clean Architecture",
			},
		},
	}
}

// LoadConfig loads reference configuration from file or returns default
func LoadConfig() (*ReferenceConfig, error) {
	configPath := getConfigPath()

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultReferenceConfig()
		if err := SaveConfig(config); err != nil {
			// If we can't save, just return default without error
			return config, nil
		}
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultReferenceConfig(), nil // Fallback to default
	}

	var config ReferenceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultReferenceConfig(), nil // Fallback to default
	}

	return &config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *ReferenceConfig) error {
	configPath := getConfigPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetReferencePath returns the path to a reference project
func (c *ReferenceConfig) GetReferencePath(templateType string) (string, error) {
	ref, exists := c.References[templateType]
	if !exists {
		return "", fmt.Errorf("unknown template type: %s", templateType)
	}

	// Convert relative paths to absolute
	if !filepath.IsAbs(ref.Path) {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		return filepath.Join(wd, ref.Path), nil
	}

	return ref.Path, nil
}

// ListTemplateTypes returns all configured template types
func (c *ReferenceConfig) ListTemplateTypes() []string {
	var types []string
	for templateType := range c.References {
		types = append(types, templateType)
	}
	return types
}

// AddReference adds a new reference project
func (c *ReferenceConfig) AddReference(templateType, path, description string) {
	if c.References == nil {
		c.References = make(map[string]ReferenceProject)
	}

	c.References[templateType] = ReferenceProject{
		Path:        path,
		Description: description,
	}
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	// Try to use XDG config directory or fallback to home
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".template-engine.json" // Fallback to current directory
		}
		configDir = filepath.Join(home, ".config")
	}

	return filepath.Join(configDir, "template-engine", "references.json")
}
