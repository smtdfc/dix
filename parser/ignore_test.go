package parser

import "testing"

func TestIgnoreMatcherExact(t *testing.T) {
	m, err := NewIgnoreMatcher([]string{"github.com/foo/bar"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.ShouldIgnore("github.com/foo/bar") {
		t.Error("expected exact match")
	}
	if m.ShouldIgnore("github.com/foo/bar/baz") {
		t.Error("exact match should not match sub-package")
	}
	if m.ShouldIgnore("github.com/foobar") {
		t.Error("exact match should not match substring")
	}
}

func TestIgnoreMatcherPrefix(t *testing.T) {
	m, err := NewIgnoreMatcher([]string{"github.com/foo/bar/..."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.ShouldIgnore("github.com/foo/bar") {
		t.Error("prefix should match parent")
	}
	if !m.ShouldIgnore("github.com/foo/bar/baz") {
		t.Error("prefix should match sub-package")
	}
	if !m.ShouldIgnore("github.com/foo/bar/baz/qux") {
		t.Error("prefix should match deep sub-package")
	}
	if m.ShouldIgnore("github.com/foo/barr") {
		t.Error("prefix should not match similar parent")
	}
}

func TestIgnoreMatcherGlob(t *testing.T) {
	m, err := NewIgnoreMatcher([]string{"github.com/foo/*bar*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.ShouldIgnore("github.com/foo/bar") {
		t.Error("glob should match bar")
	}
	if !m.ShouldIgnore("github.com/foo/foobar") {
		t.Error("glob should match foobar")
	}
	if m.ShouldIgnore("github.com/foo/baz") {
		t.Error("glob should not match baz")
	}
}

func TestIgnoreMatcherSlashStar(t *testing.T) {
	m, err := NewIgnoreMatcher([]string{"github.com/foo/bar/*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.ShouldIgnore("github.com/foo/bar") {
		t.Error("foo/bar/* should match foo/bar")
	}
	if !m.ShouldIgnore("github.com/foo/bar/baz") {
		t.Error("foo/bar/* should match foo/bar/baz")
	}
	if !m.ShouldIgnore("github.com/foo/bar/baz/qux") {
		t.Error("foo/bar/* should match deep sub-package")
	}
	if m.ShouldIgnore("github.com/foo/barr") {
		t.Error("foo/bar/* should not match foo/barr")
	}
}

func TestIgnoreMatcherExactNoFalsePositive(t *testing.T) {
	m, err := NewIgnoreMatcher([]string{"internal/mock"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.ShouldIgnore("internal/mock") {
		t.Error("exact should match internal/mock")
	}
	if m.ShouldIgnore("github.com/foo/internal/mock") {
		t.Error("exact should NOT match internal/mock as substring (no false-positive)")
	}
	if m.ShouldIgnore("internal/mocking") {
		t.Error("exact should NOT match internal/mocking")
	}
}

func TestIgnoreMatcherEmptyPatterns(t *testing.T) {
	m, err := NewIgnoreMatcher([]string{"", "   ", "github.com/foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.ShouldIgnore("github.com/foo") {
		t.Error("should still match valid pattern among empties")
	}
}

func TestShouldIgnorePkgDeprecated(t *testing.T) {
	// Giữ lại để đảm bảo hàm cũ không panic và không làm sai kết quả cơ bản
	patterns := []string{"github.com/foo/bar/..."}
	if !shouldIgnorePkg("github.com/foo/bar/baz", patterns) {
		t.Error("shouldIgnorePkg should still work")
	}
	if shouldIgnorePkg("github.com/other", patterns) {
		t.Error("shouldIgnorePkg should not match unrelated")
	}
}

// Benchmark so sánh O(N×M) cũ vs matcher mới
func BenchmarkIgnoreMatcher(b *testing.B) {
	patterns := []string{
		"github.com/foo/bar/...",
		"github.com/baz/qux/*",
		"github.com/alpha/beta",
		"internal/mock",
		"*.test.*",
	}
	m, err := NewIgnoreMatcher(patterns)
	if err != nil {
		b.Fatal(err)
	}

	pkgs := []string{
		"github.com/foo/bar",
		"github.com/foo/bar/baz",
		"github.com/foo/bar/baz/qux",
		"github.com/baz/qux",
		"github.com/baz/qux/deep",
		"github.com/alpha/beta",
		"internal/mock",
		"github.com/alpha/beta/gamma",
		"github.com/other/pkg",
		"github.com/some.test.pkg",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pkg := range pkgs {
			m.ShouldIgnore(pkg)
		}
	}
}

func BenchmarkOldShouldIgnore(b *testing.B) {
	patterns := []string{
		"github.com/foo/bar/...",
		"github.com/baz/qux/*",
		"github.com/alpha/beta",
		"internal/mock",
		"*.test.*",
	}

	pkgs := []string{
		"github.com/foo/bar",
		"github.com/foo/bar/baz",
		"github.com/foo/bar/baz/qux",
		"github.com/baz/qux",
		"github.com/baz/qux/deep",
		"github.com/alpha/beta",
		"internal/mock",
		"github.com/alpha/beta/gamma",
		"github.com/other/pkg",
		"github.com/some.test.pkg",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pkg := range pkgs {
			shouldIgnorePkg(pkg, patterns)
		}
	}
}
