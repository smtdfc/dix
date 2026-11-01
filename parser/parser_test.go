package parser

import (
	"testing"
)

func TestParseProviderLocationString(t *testing.T) {
	pErr := NewValidationErrorWithLocation("test message", "NewService", "db", "service.go", 6, 17)
	if pErr.LocationString() != "service.go:6:17" {
		t.Errorf("expected service.go:6:17, got %q", pErr.LocationString())
	}
	if pErr.Line != 6 || pErr.Column != 17 {
		t.Errorf("expected line 6 col 17, got line %d col %d", pErr.Line, pErr.Column)
	}

	// Verify LocationString on Provider
	prov := &Provider{
		File:   "service.go",
		Line:   6,
		Column: 1,
		Name:   "NewService",
	}
	if prov.LocationString() != "service.go:6:1" {
		t.Errorf("expected service.go:6:1, got %q", prov.LocationString())
	}

	// Verify LocationString on Dependency
	dep := &Dependency{
		File:   "service.go",
		Line:   6,
		Column: 17,
		Name:   "db",
	}
	if dep.LocationString() != "service.go:6:17" {
		t.Errorf("expected service.go:6:17, got %q", dep.LocationString())
	}
}

func TestParseFileLocation(t *testing.T) {
	file, line, col := parseFileLocation("main.go:42:15")
	if file != "main.go" || line != 42 || col != 15 {
		t.Errorf("expected main.go, 42, 15, got %s, %d, %d", file, line, col)
	}

	file2, line2, col2 := parseFileLocation("service.go:10")
	if file2 != "service.go" || line2 != 10 || col2 != 0 {
		t.Errorf("expected service.go, 10, 0, got %s, %d, %d", file2, line2, col2)
	}

	file3, line3, col3 := parseFileLocation("plain.go")
	if file3 != "plain.go" || line3 != 0 || col3 != 0 {
		t.Errorf("expected plain.go, 0, 0, got %s, %d, %d", file3, line3, col3)
	}
}
