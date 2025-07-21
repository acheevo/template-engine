package cmd

import (
	"github.com/acheevo/template-engine/internal/generate"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate project from template file",
	Long: `Generate a new project from an existing template schema file.

This command takes a template schema (created with 'extract') and generates
a new project with the specified parameters.

Examples:
  template-engine generate frontend-template.json --project-name "My App" --github-repo "user/my-app"
  template-engine generate api-template.json --project-name "My API" --github-repo "user/my-api"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delegate to existing generate logic
		return generate.Run()
	},
}
