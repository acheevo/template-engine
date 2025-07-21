package cmd

import (
	"fmt"
	"sort"

	"github.com/acheevo/template-engine/internal/core"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available template types",
	Long: `List all registered template types that can be used for extraction and generation.

Template types define how different kinds of projects should be processed
(file patterns to include/exclude, template variables, etc.).

Example:
  template-engine list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func runList() error {
	fmt.Println("Available template types:")
	fmt.Println()

	templates := core.ListTemplates()
	if len(templates) == 0 {
		fmt.Println("No templates registered")
		return nil
	}

	sort.Strings(templates)
	for _, templateName := range templates {
		fmt.Printf("• %s\n", templateName)
	}

	fmt.Println()
	fmt.Println("Use 'template-engine new <type> <name> <repo>' to create a project")

	return nil
}
