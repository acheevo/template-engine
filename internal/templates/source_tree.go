package templates

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/acheevo/template-engine/internal/core"
	"github.com/acheevo/template-engine/internal/envparser"
)

// extractSourceTree packages regular files, with paths and substitutions provided by the template.
func extractSourceTree(
	sourceDir string, schema *core.TemplateSchema,
	skipPath func(string, bool) bool, mappingsFor func(string) []core.Mapping,
) (*core.TemplateSchema, error) {
	err := filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || path == sourceDir {
			return walkErr
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if skipPath(rel, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content, compressed, err := core.CompressContent(string(raw))
		if err != nil {
			return err
		}
		mappings := mappingsFor(rel)
		schema.Files = append(schema.Files, core.FileSpec{
			Path: rel, Content: content, Compressed: compressed, Template: len(mappings) > 0,
			Size: int64(len(raw)), Hash: core.CalculateContentHash(string(raw)), Mappings: mappings,
		})
		if rel == envExampleFile {
			schema.EnvConfig = envparser.ParseEnvExample(string(raw))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := core.ValidateSchema(schema); err != nil {
		return nil, err
	}
	var digest strings.Builder
	digest.WriteString(schema.Name + schema.Type + schema.Version)
	for _, file := range schema.Files {
		digest.WriteString(file.Path + file.Hash)
	}
	schema.Hash = core.CalculateContentHash(digest.String())
	return schema, nil
}
