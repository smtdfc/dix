package cmd

import (
	"strings"
	"testing"

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
