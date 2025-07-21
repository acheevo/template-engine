package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "template-engine",
	Short: "Generate projects from templates",
	Long: `Template Engine - Generate projects from templates

Quick Start:
  template-engine new frontend "My React App" "user/my-app"
  template-engine new go-api "My API" "user/my-api"
  template-engine new --interactive

Advanced Usage:
  template-engine extract <source-dir> --type <template-type> [-o output.json]
  template-engine generate <template.json> --project-name <name> --github-repo <repo>
  template-engine list [--verbose]`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(extractCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(configCmd)
}
