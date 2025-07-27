package extract

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/acheevo/template-engine/internal/core"
)

func RunWithParams(sourceDir, outputFile, templateType string) error {
	if templateType == "" {
		return fmt.Errorf("--type flag is required. Available types: %v", core.ListTemplates())
	}

	fmt.Printf("Extracting %s template from %s to %s\n", templateType, sourceDir, outputFile)

	return extract(sourceDir, outputFile, templateType)
}

func Run() error {
	args := os.Args[2:]

	if len(args) == 0 {
		return fmt.Errorf("usage: template-engine extract <source-dir> --type <template-type> [flags]")
	}

	sourceDir := args[0]
	outputFile := "template.json"
	templateType := ""

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", args[i])
			}
			outputFile = args[i+1]
			i++
		case "--type":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", args[i])
			}
			templateType = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	return RunWithParams(sourceDir, outputFile, templateType)
}

func extract(sourceDir, outputFile, templateType string) error {
	// Check if source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// Get template type from registry
	template, err := core.GetTemplate(templateType)
	if err != nil {
		return fmt.Errorf("failed to get template type: %w", err)
	}

	// Extract using the specific template type
	schema, err := template.Extract(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to extract template: %w", err)
	}

	// Save to file
	err = saveSchemaToFile(schema, outputFile)
	if err != nil {
		return fmt.Errorf("failed to save template to file: %w", err)
	}

	fmt.Printf("Template extracted successfully to %s\n", outputFile)
	fmt.Printf("Template type: %s\n", schema.Type)
	fmt.Printf("Found %d files (%d templated)\n",
		len(schema.Files),
		countTemplatedFiles(schema.Files))
	fmt.Printf("Total size: %s\n", formatSize(calculateTotalSize(schema.Files)))

	return nil
}

func saveSchemaToFile(schema *core.TemplateSchema, filename string) error {
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0o600)
}

func countTemplatedFiles(files []core.FileSpec) int {
	count := 0
	for _, file := range files {
		if file.Template {
			count++
		}
	}
	return count
}

func calculateTotalSize(files []core.FileSpec) int64 {
	var total int64
	for _, file := range files {
		total += file.Size
	}
	return total
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
