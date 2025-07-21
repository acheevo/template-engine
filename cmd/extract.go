package cmd

import (
	"github.com/acheevo/template-engine/internal/extract"
	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract a template from source directory",
	Long: `Extract a template schema from an existing project directory.
	
This command analyzes a source project and creates a reusable template
that can be used to generate similar projects.

Examples:
  template-engine extract ../my-frontend --type frontend -o frontend-template.json
  template-engine extract ../my-api --type go-api -o api-template.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delegate to existing extract logic
		return extract.Run()
	},
}
