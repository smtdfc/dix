package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/smtdfc/dix/generator"
	"github.com/smtdfc/dix/parser"
)

func TestFormatDixErrorPrintsAReadableCauseChain(t *testing.T) {
	inner := parser.NewValidationError("missing dependency", "NewApp", "repo", "app.go:12")
	err := generator.NewGenerateError(generator.ErrorGraphBuild, "failed to build provider graph", "NewApp", "repo", inner)

	got := formatDixError(err)
	if !strings.Contains(got, "generator/graph_build: failed to build provider graph") {
		t.Fatalf("formatted error = %q", got)
	}
	if !strings.Contains(got, "caused by: missing dependency [fn=NewApp] [field=repo] (app.go:12)") {
		t.Fatalf("formatted error does not contain the source location: %q", got)
	}
	if strings.Contains(got, "failed to build provider graph: missing dependency") {
		t.Fatalf("formatted error still nests messages inline: %q", got)
	}
}

func TestFormatDiagnosticWithSourceSnippet(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "service.go")
	sourceContent := `package app

// @Injectable
func NewMessage(repo *Repo) string {
	return "hello"
}
`
	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	genErr := generator.NewGenerateErrorWithLocation(
		generator.ErrorDependencyResolve,
		"provider not found for dependency",
		"NewMessage",
		"repo (type of *Repo in package app)",
		sourceFile,
		4,
		17,
		nil,
	)

	diag := FormatDiagnostic(genErr)
	t.Log("\n" + diag)

	if !strings.Contains(diag, "[Dix Error]") {
		t.Errorf("diagnostic missing badge: %s", diag)
	}
	if !strings.Contains(diag, "Dependency Resolution Error: provider not found for dependency") {
		t.Errorf("diagnostic missing title and message: %s", diag)
	}
	if !strings.Contains(diag, sourceFile+":4:17") {
		t.Errorf("diagnostic missing file location: %s", diag)
	}
	if !strings.Contains(diag, "func NewMessage(repo *Repo) string {") {
		t.Errorf("diagnostic missing source line: %s", diag)
	}
	if !strings.Contains(diag, "^^^^") {
		t.Errorf("diagnostic missing caret underline: %s", diag)
	}
	if !strings.Contains(diag, "Provider:   NewMessage") {
		t.Errorf("diagnostic missing provider: %s", diag)
	}
	if !strings.Contains(diag, "Dependency: repo (type of *Repo in package app)") {
		t.Errorf("diagnostic missing dependency: %s", diag)
	}
	if !strings.Contains(diag, "💡 Hint:") {
		t.Errorf("diagnostic missing hint: %s", diag)
	}
}

func TestFormatDiagnosticCircularDependency(t *testing.T) {
	color.NoColor = true
	defer func() { color.NoColor = false }()

	err := generator.NewGenerateErrorWithLocation(
		generator.ErrorDependencyResolve,
		"circular dependency detected",
		"NewServiceA",
		"",
		"app/service.go",
		25,
		1,
		nil,
	)

	diag := FormatDiagnostic(err)
	t.Log("\n" + diag)
	if !strings.Contains(diag, "circular dependency detected") {
		t.Errorf("expected circular dependency message in %s", diag)
	}
	if !strings.Contains(diag, "app/service.go:25:1") {
		t.Errorf("expected location in %s", diag)
	}
	if !strings.Contains(diag, "💡 Hint:") || !strings.Contains(diag, "cycle") {
		t.Errorf("expected cycle hint in %s", diag)
	}
}
