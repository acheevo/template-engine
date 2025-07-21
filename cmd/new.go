package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/acheevo/template-engine/internal/config"
	"github.com/acheevo/template-engine/sdk"
	"github.com/spf13/cobra"
)

var interactive bool

var newCmd = &cobra.Command{
	Use:   "new [type] [project-name] [github-repo] [output-dir]",
	Short: "Quick project generation from reference projects",
	Long: `Generate a new project by extracting a template from a reference project
and immediately generating from it. This is the fastest way to create new projects.

The command looks for reference projects in sibling directories:
- frontend: ../frontend-template
- go-api:   ../api-template

Examples:
  template-engine new frontend "My React App" "user/my-app"
  template-engine new go-api "My API Service" "user/my-api"
  template-engine new --interactive`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if interactive {
			return runInteractiveNew()
		}

		if len(args) < 3 {
			return fmt.Errorf("usage: template-engine new <template-type> <project-name> <github-repo> [output-dir]")
		}

		templateType := args[0]
		projectName := args[1]
		githubRepo := args[2]

		outputDir := "./" + strings.ToLower(strings.ReplaceAll(projectName, " ", "-"))
		if len(args) > 3 {
			outputDir = args[3]
		}

		return runNew(templateType, projectName, githubRepo, outputDir)
	},
}

func init() {
	newCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive project creation mode")
}

func runNew(templateType, projectName, githubRepo, outputDir string) error {
	// Load reference configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get reference project path
	referenceDir, err := cfg.GetReferencePath(templateType)
	if err != nil {
		return err
	}

	// Check if reference project exists
	if _, err := os.Stat(referenceDir); os.IsNotExist(err) {
		return fmt.Errorf("reference project not found: %s. Make sure you have the reference project available", referenceDir)
	}

	fmt.Printf("🚀 Creating %s project...\n", templateType)
	fmt.Printf("   Reference: %s\n", referenceDir)
	fmt.Printf("   Name: %s\n", projectName)
	fmt.Printf("   Repo: %s\n", githubRepo)
	fmt.Printf("   Output: %s\n", outputDir)
	fmt.Println()

	// Use SDK to extract and generate
	client := sdk.New()

	err = client.ExtractAndGenerate(context.Background(), referenceDir, templateType, projectName, githubRepo, outputDir)
	if err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Print success message and next steps
	fmt.Println()
	fmt.Printf("✨ Project created successfully!\n")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", filepath.Base(outputDir))

	switch templateType {
	case "frontend":
		fmt.Println("  npm install")
		fmt.Println("  npm run dev")
	case "go-api", "api":
		fmt.Println("  go mod tidy")
		fmt.Println("  make run")
	}

	return nil
}

func runInteractiveNew() error {
	fmt.Println("🎯 Interactive Project Generator")
	fmt.Println()

	// Load configuration to get available template types
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	templateTypes := cfg.ListTemplateTypes()
	if len(templateTypes) == 0 {
		return fmt.Errorf("no template types configured")
	}

	// Template type selection
	fmt.Println("Select template type:")
	for i, templateType := range templateTypes {
		ref := cfg.References[templateType]
		fmt.Printf("%d. %s - %s\n", i+1, templateType, ref.Description)
	}
	fmt.Printf("Enter choice (1-%d): ", len(templateTypes))

	var choice int
	if _, err := fmt.Scanln(&choice); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	if choice < 1 || choice > len(templateTypes) {
		return fmt.Errorf("invalid choice")
	}

	templateType := templateTypes[choice-1]

	// Project details
	var projectName, githubRepo string

	fmt.Print("Project name: ")
	if _, err := fmt.Scanln(&projectName); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	fmt.Print("GitHub repo (user/repo-name): ")
	if _, err := fmt.Scanln(&githubRepo); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	outputDir := "./" + strings.ToLower(strings.ReplaceAll(projectName, " ", "-"))

	return runNew(templateType, projectName, githubRepo, outputDir)
}
