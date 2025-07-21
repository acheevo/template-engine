package sdk

import "fmt"

// ErrorType represents different categories of SDK errors
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeExtraction   ErrorType = "extraction"
	ErrorTypeGeneration   ErrorType = "generation"
	ErrorTypeTemplateType ErrorType = "template_type"
	ErrorTypeFileSystem   ErrorType = "filesystem"
	ErrorTypeSchema       ErrorType = "schema"
)

// SDKError provides structured error information for SDK operations
type SDKError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Operation  string    `json:"operation"`
	Details    string    `json:"details,omitempty"`
	Underlying error     `json:"-"`
}

func (e *SDKError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s error in %s: %s (%s)", e.Type, e.Operation, e.Message, e.Details)
	}
	return fmt.Sprintf("%s error in %s: %s", e.Type, e.Operation, e.Message)
}

func (e *SDKError) Unwrap() error {
	return e.Underlying
}

// newValidationError creates a validation error
func newValidationError(operation, message, details string) *SDKError {
	return &SDKError{
		Type:      ErrorTypeValidation,
		Operation: operation,
		Message:   message,
		Details:   details,
	}
}

// newExtractionError creates an extraction error
func newExtractionError(operation, message string, underlying error) *SDKError {
	return &SDKError{
		Type:       ErrorTypeExtraction,
		Operation:  operation,
		Message:    message,
		Underlying: underlying,
	}
}

// newGenerationError creates a generation error
func newGenerationError(operation, message string, underlying error) *SDKError {
	return &SDKError{
		Type:       ErrorTypeGeneration,
		Operation:  operation,
		Message:    message,
		Underlying: underlying,
	}
}

// newTemplateTypeError creates a template type error
func newTemplateTypeError(operation, templateType string) *SDKError {
	return &SDKError{
		Type:      ErrorTypeTemplateType,
		Operation: operation,
		Message:   fmt.Sprintf("unknown template type: %s", templateType),
		Details:   "Use ListTemplates() to see available types",
	}
}

// newFileSystemError creates a filesystem error
func newFileSystemError(operation, message string, underlying error) *SDKError {
	return &SDKError{
		Type:       ErrorTypeFileSystem,
		Operation:  operation,
		Message:    message,
		Underlying: underlying,
	}
}

// newSchemaError creates a schema error
func newSchemaError(operation, message string, underlying error) *SDKError {
	return &SDKError{
		Type:       ErrorTypeSchema,
		Operation:  operation,
		Message:    message,
		Underlying: underlying,
	}
}
