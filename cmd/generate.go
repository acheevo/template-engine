package cmd

import (
	"github.com/acheevo/template-engine/internal/generate"
	"github.com/spf13/cobra"
)

var (
	generateProjectName string
	generateGithubRepo  string
	generateOutputDir   string
)

var generateCmd = &cobra.Command{
	Use:   "generate <template-file>",
	Short: "Generate project from template file",
	Long: `Generate a new project from an existing template schema file.

This command takes a template schema (created with 'extract') and generates
a new project with the specified parameters.

Examples:
  template-engine generate frontend-template.json --project-name "My App" --github-repo "user/my-app"
  template-engine generate api-template.json --project-name "My API" --github-repo "user/my-api"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateFile := args[0]
		return generate.RunWithParams(templateFile, generateOutputDir, generateProjectName, generateGithubRepo)
	},
}

func init() {
	generateCmd.Flags().StringVar(&generateProjectName, "project-name", "", "Name of the project (required)")
	generateCmd.Flags().StringVar(&generateGithubRepo, "github-repo", "",
		"GitHub repository (e.g., username/repo-name) (required)")
	generateCmd.Flags().StringVar(&generateOutputDir, "output-dir", "./", "Output directory for generated project")
	_ = generateCmd.MarkFlagRequired("project-name")
	_ = generateCmd.MarkFlagRequired("github-repo")
}
