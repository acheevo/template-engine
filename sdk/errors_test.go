package sdk

import (
	"errors"
	"testing"
)

func TestSDKError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *SDKError
		expected string
	}{
		{
			name: "error with details",
			err: &SDKError{
				Type:      ErrorTypeValidation,
				Operation: "TestOperation",
				Message:   "test message",
				Details:   "additional details",
			},
			expected: "validation error in TestOperation: test message (additional details)",
		},
		{
			name: "error without details",
			err: &SDKError{
				Type:      ErrorTypeExtraction,
				Operation: "Extract",
				Message:   "extraction failed",
			},
			expected: "extraction error in Extract: extraction failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("SDKError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSDKError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &SDKError{
		Type:       ErrorTypeFileSystem,
		Operation:  "TestOperation",
		Message:    "test message",
		Underlying: underlying,
	}

	if unwrapped := err.Unwrap(); unwrapped != underlying {
		t.Errorf("SDKError.Unwrap() = %v, want %v", unwrapped, underlying)
	}
}

func TestErrorConstructors(t *testing.T) {
	underlying := errors.New("underlying error")

	tests := []struct {
		name         string
		constructor  func() *SDKError
		expectedType ErrorType
		expectedOp   string
	}{
		{
			name: "newValidationError",
			constructor: func() *SDKError {
				return newValidationError("TestOp", "validation failed", "details")
			},
			expectedType: ErrorTypeValidation,
			expectedOp:   "TestOp",
		},
		{
			name: "newExtractionError",
			constructor: func() *SDKError {
				return newExtractionError("ExtractOp", "extraction failed", underlying)
			},
			expectedType: ErrorTypeExtraction,
			expectedOp:   "ExtractOp",
		},
		{
			name: "newGenerationError",
			constructor: func() *SDKError {
				return newGenerationError("GenerateOp", "generation failed", underlying)
			},
			expectedType: ErrorTypeGeneration,
			expectedOp:   "GenerateOp",
		},
		{
			name: "newTemplateTypeError",
			constructor: func() *SDKError {
				return newTemplateTypeError("TestOp", "unknown-type")
			},
			expectedType: ErrorTypeTemplateType,
			expectedOp:   "TestOp",
		},
		{
			name: "newFileSystemError",
			constructor: func() *SDKError {
				return newFileSystemError("FileOp", "file system error", underlying)
			},
			expectedType: ErrorTypeFileSystem,
			expectedOp:   "FileOp",
		},
		{
			name: "newSchemaError",
			constructor: func() *SDKError {
				return newSchemaError("SchemaOp", "schema error", underlying)
			},
			expectedType: ErrorTypeSchema,
			expectedOp:   "SchemaOp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()

			if err.Type != tt.expectedType {
				t.Errorf("Expected error type %v, got %v", tt.expectedType, err.Type)
			}

			if err.Operation != tt.expectedOp {
				t.Errorf("Expected operation %v, got %v", tt.expectedOp, err.Operation)
			}

			if err.Message == "" {
				t.Error("Expected non-empty message")
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	expectedTypes := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeExtraction,
		ErrorTypeGeneration,
		ErrorTypeTemplateType,
		ErrorTypeFileSystem,
		ErrorTypeSchema,
	}

	expectedValues := []string{
		"validation",
		"extraction",
		"generation",
		"template_type",
		"filesystem",
		"schema",
	}

	for i, expectedType := range expectedTypes {
		if string(expectedType) != expectedValues[i] {
			t.Errorf("Expected error type %v to have value %v, got %v",
				expectedType, expectedValues[i], string(expectedType))
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	underlying := errors.New("original error")

	err := newFileSystemError("TestOp", "file system error", underlying)

	// Test that errors.Is works
	if !errors.Is(err, underlying) {
		t.Error("Expected errors.Is to return true for underlying error")
	}

	// Test that errors.Unwrap works
	if unwrapped := errors.Unwrap(err); unwrapped != underlying {
		t.Errorf("Expected errors.Unwrap to return underlying error, got %v", unwrapped)
	}
}
