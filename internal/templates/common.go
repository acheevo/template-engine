package templates

import (
	"path/filepath"
	"strings"
)

// Common template file names
const (
	ReadmeFile = "README.md"
)

// shouldSkipCommon contains common logic for skipping files during template extraction
func shouldSkipCommon(path string, skipDirs []string) bool {
	// Always include .github directories and their contents
	if strings.Contains(path, ".github") {
		return false
	}

	// Skip .git directory and all its contents
	if strings.Contains(path, ".git") && !strings.Contains(path, ".github") {
		return true
	}

	baseName := filepath.Base(path)

	// Skip other hidden files/directories (starting with .) except .github
	if strings.HasPrefix(baseName, ".") && baseName != ".github" && !strings.Contains(path, ".github") {
		return true
	}

	// Skip specific directories (check if directory appears anywhere in path)
	for _, dir := range skipDirs {
		if strings.Contains(path, string(filepath.Separator)+dir+string(filepath.Separator)) ||
			strings.HasSuffix(path, string(filepath.Separator)+dir) ||
			strings.HasPrefix(path, dir+string(filepath.Separator)) {
			return true
		}
	}

	// Skip file patterns
	if strings.HasSuffix(baseName, ".log") {
		return true
	}

	return false
}
