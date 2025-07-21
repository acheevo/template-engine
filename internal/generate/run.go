package generate

import (
	"fmt"
	"os"
)

func Run() error {
	args := os.Args[2:]

	if len(args) == 0 {
		return fmt.Errorf("usage: template-engine generate <template.json> [flags]")
	}

	templateFile := args[0]
	outputDir := "./"
	projectName := ""
	githubRepo := ""

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--project-name":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", args[i])
			}
			projectName = args[i+1]
			i++
		case "--github-repo":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", args[i])
			}
			githubRepo = args[i+1]
			i++
		case "--output-dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", args[i])
			}
			outputDir = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	if projectName == "" {
		return fmt.Errorf("--project-name is required")
	}

	if githubRepo == "" {
		return fmt.Errorf("--github-repo is required")
	}

	fmt.Printf("Generating project from %s\n", templateFile)
	fmt.Printf("Project name: %s\n", projectName)
	fmt.Printf("GitHub repo: %s\n", githubRepo)
	fmt.Printf("Output dir: %s\n", outputDir)

	return generate(templateFile, outputDir, projectName, githubRepo)
}

func generate(templateFile, outputDir, projectName, githubRepo string) error {
	// Check if template file exists
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		return fmt.Errorf("template file does not exist: %s", templateFile)
	}

	// Check if output directory already exists
	if _, err := os.Stat(outputDir); err == nil {
		return fmt.Errorf("output directory already exists: %s", outputDir)
	}

	// Create generator
	generator, err := NewGenerator(templateFile, outputDir, projectName, githubRepo)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	// Generate project
	if err := generator.Generate(); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Print summary
	generator.PrintSummary()

	return nil
}
