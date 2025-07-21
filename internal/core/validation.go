package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// ValidateSchema validates a template schema for integrity and completeness
func ValidateSchema(schema *TemplateSchema) error {
	if err := validateBasicFields(schema); err != nil {
		return err
	}

	if err := validateSchemaVariables(schema); err != nil {
		return err
	}

	return validateSchemaFiles(schema)
}

// validateBasicFields validates the basic required fields
func validateBasicFields(schema *TemplateSchema) error {
	if schema.Name == "" {
		return fmt.Errorf("schema name is required")
	}

	if schema.Type == "" {
		return fmt.Errorf("schema type is required")
	}

	if schema.Version == "" {
		return fmt.Errorf("schema version is required")
	}

	return nil
}

// validateSchemaVariables validates the variables section
func validateSchemaVariables(schema *TemplateSchema) error {
	if schema.Variables == nil {
		return fmt.Errorf("schema variables is required")
	}

	for name, variable := range schema.Variables {
		if variable.Type == "" {
			return fmt.Errorf("variable %s must have a type", name)
		}
	}

	return nil
}

// validateSchemaFiles validates the files section
func validateSchemaFiles(schema *TemplateSchema) error {
	if len(schema.Files) == 0 {
		return fmt.Errorf("schema must contain at least one file")
	}

	for i, file := range schema.Files {
		if err := validateFileSpec(file, i); err != nil {
			return err
		}
	}

	return nil
}

// validateFileSpec validates a single file specification
func validateFileSpec(file FileSpec, index int) error {
	if file.Path == "" {
		return fmt.Errorf("file %d must have a path", index)
	}

	if file.Content == "" {
		return fmt.Errorf("file %s must have content", file.Path)
	}

	return validateFileHash(file)
}

// validateFileHash validates the hash of a file if present
func validateFileHash(file FileSpec) error {
	if file.Hash == "" {
		return nil
	}

	content := file.Content
	if file.Compressed {
		decompressed, err := DecompressContent(file.Content, file.Compressed)
		if err != nil {
			return fmt.Errorf("file %s failed to decompress for validation: %w", file.Path, err)
		}
		content = decompressed
	}

	calculatedHash := CalculateContentHash(content)
	if file.Hash != calculatedHash {
		return fmt.Errorf("file %s hash mismatch: expected %s, got %s",
			file.Path, file.Hash, calculatedHash)
	}

	return nil
}

// ValidateVariables validates that all required variables are provided
func ValidateVariables(schema *TemplateSchema, variables *TemplateVariables) error {
	for name, variable := range schema.Variables {
		if variable.Required {
			switch name {
			case "ProjectName":
				if variables.ProjectName == "" {
					return fmt.Errorf("ProjectName is required")
				}
			case "GitHubRepo":
				if variables.GitHubRepo == "" {
					return fmt.Errorf("GitHubRepo is required")
				}
			case "Author":
				if variables.Author == "" && variable.Default == "" {
					return fmt.Errorf("author is required")
				}
			case "Description":
				if variables.Description == "" && variable.Default == "" {
					return fmt.Errorf("description is required")
				}
			}
		}
	}

	return nil
}

// CalculateContentHash calculates SHA256 hash of content
func CalculateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
