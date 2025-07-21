package cmd

import (
	"fmt"
	"sort"

	"github.com/acheevo/template-engine/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage reference project configuration",
	Long: `Configure where reference projects are located for template generation.

Reference projects are existing projects that serve as templates for generating
new projects. You can add, list, or remove reference project configurations.

Examples:
  template-engine config list
  template-engine config add my-template /path/to/template "My custom template"
  template-engine config remove my-template`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured reference projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigList()
	},
}

var configAddCmd = &cobra.Command{
	Use:   "add [template-type] [path] [description]",
	Short: "Add a new reference project",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigAdd(args[0], args[1], args[2])
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove [template-type]",
	Short: "Remove a reference project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigRemove(args[0])
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
}

func runConfigList() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(cfg.References) == 0 {
		fmt.Println("No reference projects configured")
		return nil
	}

	fmt.Println("Configured reference projects:")
	fmt.Println()

	// Sort template types for consistent output
	var types []string
	for templateType := range cfg.References {
		types = append(types, templateType)
	}
	sort.Strings(types)

	for _, templateType := range types {
		ref := cfg.References[templateType]
		fmt.Printf("• %s\n", templateType)
		fmt.Printf("  Path: %s\n", ref.Path)
		fmt.Printf("  Description: %s\n", ref.Description)
		if ref.Version != "" {
			fmt.Printf("  Version: %s\n", ref.Version)
		}
		fmt.Println()
	}

	return nil
}

func runConfigAdd(templateType, path, description string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg.AddReference(templateType, path, description)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Added reference project '%s' at %s\n", templateType, path)
	return nil
}

func runConfigRemove(templateType string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if _, exists := cfg.References[templateType]; !exists {
		return fmt.Errorf("template type '%s' not found", templateType)
	}

	delete(cfg.References, templateType)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Removed reference project '%s'\n", templateType)
	return nil
}
