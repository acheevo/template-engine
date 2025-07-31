package sdk

// This file previously contained duplicate template implementations.
//
// The SDK now uses the canonical template implementations from internal/templates/
// through the global template registry system. This eliminates code duplication
// and ensures the SDK has access to all the sophisticated features like:
// - Advanced file mapping logic
// - Content compression
// - Environment variable parsing
// - Complex template variable handling
//
// Template operations are split into two categories:
//
// 1. Template Type Operations (use global registry):
//    - ListTemplates() - Available template types
//    - GetTemplateInfo() - Template metadata
//    - Extract() - Extract schema from source directory
//
// 2. Template Schema Operations (use client cache):
//    - RegisterTemplate() - Register pre-extracted schema files
//    - Generate() - Generate from registered schemas
