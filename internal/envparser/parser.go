package envparser

import (
	"bufio"
	"strings"

	"github.com/acheevo/template-engine/internal/core"
)

// ParseEnvExample parses a .env.example file and returns environment variables
func ParseEnvExample(content string) []core.EnvVariable {
	var envVars []core.EnvVariable
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentDescription string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			currentDescription = ""
			continue
		}

		// Handle comment lines for descriptions
		if strings.HasPrefix(line, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if comment != "" {
				currentDescription = comment
			}
			continue
		}

		// Handle environment variable lines
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				example := strings.TrimSpace(parts[1])

				envVar := core.EnvVariable{
					Name:        name,
					Description: currentDescription,
					Example:     example,
				}

				envVars = append(envVars, envVar)
				currentDescription = "" // Reset description after use
			}
		}
	}

	return envVars
}
