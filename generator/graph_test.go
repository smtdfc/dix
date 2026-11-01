package generator

import (
	"errors"
	"testing"

	"github.com/smtdfc/dix/parser"
)

func TestBuildGraphMissingDependencyCapturesLocation(t *testing.T) {
	root := &parser.Provider{
		Name: "NewApp",
		File: "app.go",
		Line: 10,
		Return: &parser.ReturnValue{
			Type: &parser.TypeInfo{Name: "App", Pkg: "main"},
		},
		Deps: []*parser.Dependency{
			{
				Name:   "service",
				File:   "app.go",
				Line:   10,
				Column: 13,
				Type:   &parser.TypeInfo{Name: "Service", Pkg: "main", IsPointer: true},
			},
		},
	}

	providerMap := ProviderMap{}

	_, err := BuildGraph(root, providerMap)
	if err == nil {
		t.Fatalf("expected error for missing dependency, got nil")
	}

	var genErr *GenerateError
	if !errors.As(err, &genErr) {
		t.Fatalf("expected GenerateError, got %T", err)
	}

	if genErr.Kind != ErrorDependencyResolve {
		t.Errorf("expected ErrorDependencyResolve, got %v", genErr.Kind)
	}
	if genErr.File != "app.go" {
		t.Errorf("expected File app.go, got %q", genErr.File)
	}
	if genErr.Line != 10 {
		t.Errorf("expected Line 10, got %d", genErr.Line)
	}
	if genErr.Column != 13 {
		t.Errorf("expected Column 13, got %d", genErr.Column)
	}
	if genErr.LocationString() != "app.go:10:13" {
		t.Errorf("expected LocationString app.go:10:13, got %q", genErr.LocationString())
	}
}

func TestBuildGraphCircularDependencyCapturesLocation(t *testing.T) {
	nodeA := &parser.Provider{
		Name:   "NewA",
		File:   "a.go",
		Line:   15,
		Column: 1,
		Return: &parser.ReturnValue{
			Type: &parser.TypeInfo{Name: "A", Pkg: "main", IsPointer: true},
		},
	}
	nodeB := &parser.Provider{
		Name:   "NewB",
		File:   "b.go",
		Line:   20,
		Column: 1,
		Return: &parser.ReturnValue{
			Type: &parser.TypeInfo{Name: "B", Pkg: "main", IsPointer: true},
		},
	}

	nodeA.Deps = []*parser.Dependency{
		{
			Name: "b",
			Type: nodeB.Return.Type,
		},
	}
	nodeB.Deps = []*parser.Dependency{
		{
			Name: "a",
			Type: nodeA.Return.Type,
		},
	}

	providerMap := ProviderMap{
		nodeA.Return.Type.Signature(): nodeA,
		nodeB.Return.Type.Signature(): nodeB,
	}

	graph, err := BuildGraph(nodeA, providerMap)
	if err != nil {
		t.Fatalf("BuildGraph returned unexpected error: %v", err)
	}

	_, sortErr := graph.Sort()
	if sortErr == nil {
		t.Fatalf("expected circular dependency error, got nil")
	}

	var genErr *GenerateError
	if !errors.As(sortErr, &genErr) {
		t.Fatalf("expected GenerateError, got %T", sortErr)
	}

	if genErr.Kind != ErrorDependencyResolve {
		t.Errorf("expected ErrorDependencyResolve, got %v", genErr.Kind)
	}
	if genErr.Line == 0 {
		t.Errorf("expected non-zero line number for circular dependency error")
	}
	if genErr.File != "a.go" && genErr.File != "b.go" {
		t.Errorf("expected File a.go or b.go, got %q", genErr.File)
	}
}
