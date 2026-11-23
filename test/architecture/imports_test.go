package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/youenchene/EnglishCraft/"

func TestHexagonalImportBoundaries(t *testing.T) {
	root := projectRoot(t)
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("parse %s: %v", path, err)
			return nil
		}
		relative, _ := filepath.Rel(root, path)
		for _, imported := range parsed.Imports {
			assertAllowedImport(t, filepath.ToSlash(relative), importPath(t, imported))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertAllowedImport(t *testing.T, source, imported string) {
	t.Helper()
	if strings.Contains(source, "/domain/") && strings.Contains(strings.Split(imported, "/")[0], ".") && !strings.HasPrefix(imported, module+"internal/shared/domain") {
		t.Errorf("domain external dependency: %s imports %s", source, imported)
	}
	if !strings.HasPrefix(imported, module) {
		return
	}
	switch {
	case strings.Contains(source, "/domain/"):
		if !strings.Contains(imported, "/internal/shared/domain") {
			t.Errorf("domain purity violation: %s imports %s", source, imported)
		}
	case strings.Contains(source, "/application/"):
		if strings.Contains(imported, "/adapters/") || strings.Contains(imported, "/platform/") || strings.Contains(imported, "/bootstrap") {
			t.Errorf("application boundary violation: %s imports %s", source, imported)
		}
	}
}

func importPath(t *testing.T, imported *ast.ImportSpec) string {
	t.Helper()
	value, err := strconv.Unquote(imported.Path.Value)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func projectRoot(t *testing.T) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(workingDirectory, "..", ".."))
}
