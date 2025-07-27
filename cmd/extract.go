package cmd

import (
	"github.com/acheevo/template-engine/internal/extract"
	"github.com/spf13/cobra"
)

var (
	extractOutputFile string
	extractType       string
)

var extractCmd = &cobra.Command{
	Use:   "extract <source-dir>",
	Short: "Extract a template from source directory",
	Long: `Extract a template schema from an existing project directory.
	
This command analyzes a source project and creates a reusable template
that can be used to generate similar projects.

Examples:
  template-engine extract ../my-frontend --type frontend -o frontend-template.json
  template-engine extract ../my-api --type go-api -o api-template.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceDir := args[0]
		return extract.RunWithParams(sourceDir, extractOutputFile, extractType)
	},
}

func init() {
	extractCmd.Flags().StringVarP(&extractOutputFile, "output", "o", "template.json",
		"Output file for the extracted template")
	extractCmd.Flags().StringVar(&extractType, "type", "", "Template type (required)")
	_ = extractCmd.MarkFlagRequired("type") // Error is not critical for flag registration
}
