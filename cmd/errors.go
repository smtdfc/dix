package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/smtdfc/dix/generator"
	"github.com/smtdfc/dix/parser"
)

func fatalDixError(err error) {
	if err == nil {
		return
	}

	fmt.Fprint(os.Stderr, FormatDiagnostic(err))
	os.Exit(1)
}

func formatDixError(err error) string {
	lines := make([]string, 0, 3)
	for current := err; current != nil; current = errors.Unwrap(current) {
		var parseErr *parser.ParseError
		if errors.As(current, &parseErr) && parseErr == current {
			lines = append(lines, parseErr.Error())
			continue
		}
		var genErr *generator.GenerateError
		if errors.As(current, &genErr) && genErr == current {
			lines = append(lines, genErr.Error())
			continue
		}
		lines = append(lines, current.Error())
	}
	for i := range lines {
		if i > 0 {
			lines[i] = strings.Repeat("  ", i) + "caused by: " + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

// FormatDiagnostic returns a user-friendly, beautifully formatted error diagnostic.
func FormatDiagnostic(err error) string {
	if err == nil {
		return ""
	}

	// Unpack causes
	var chain []error
	for cur := err; cur != nil; cur = errors.Unwrap(cur) {
		chain = append(chain, cur)
	}

	var bestGenErr *generator.GenerateError
	var bestParseErr *parser.ParseError

	for _, cur := range chain {
		var gErr *generator.GenerateError
		if errors.As(cur, &gErr) {
			if bestGenErr == nil || (bestGenErr.File == "" && gErr.File != "") {
				bestGenErr = gErr
			}
		}
		var pErr *parser.ParseError
		if errors.As(cur, &pErr) {
			if bestParseErr == nil || (bestParseErr.File == "" && pErr.File != "") {
				bestParseErr = pErr
			}
		}
	}

	var title, message, provider, dependsOn, field, file string
	var line, col int

	if bestParseErr != nil && (bestGenErr == nil || (bestParseErr.File != "" && bestGenErr.File == "")) {
		title = formatParseErrorKind(bestParseErr.Kind)
		message = bestParseErr.Message
		provider = bestParseErr.Function
		field = bestParseErr.Field
		file = bestParseErr.File
		line = bestParseErr.Line
		col = bestParseErr.Column
	} else if bestGenErr != nil {
		title = formatGenerateErrorKind(bestGenErr.Kind)
		message = bestGenErr.Message
		provider = bestGenErr.Provider
		dependsOn = strings.TrimSpace(bestGenErr.DependsOn)
		file = bestGenErr.File
		line = bestGenErr.Line
		col = bestGenErr.Column
	} else {
		title = "Dix Error"
		message = err.Error()
	}

	// Parse file location if file contains path:line or path:line:col
	if line == 0 && file != "" {
		file, line, col = parseLocationString(file)
	}

	var b strings.Builder

	// Header badge and title
	badge := colorize("\033[1;31m", "[Dix Error]")
	titleStr := colorize("\033[1m", title)
	if message != "" {
		b.WriteString(fmt.Sprintf("%s %s: %s\n\n", badge, titleStr, message))
	} else {
		b.WriteString(fmt.Sprintf("%s %s\n\n", badge, titleStr))
	}

	// Source snippet
	highlightWord := field
	if highlightWord == "" && dependsOn != "" {
		parts := strings.Fields(dependsOn)
		if len(parts) > 0 {
			highlightWord = parts[0]
		}
	}

	snippet := renderSourceSnippet(file, line, col, highlightWord)
	if snippet != "" {
		b.WriteString(snippet)
		b.WriteString("\n")
	} else if file != "" {
		locStr := file
		if line > 0 {
			if col > 0 {
				locStr = fmt.Sprintf("%s:%d:%d", file, line, col)
			} else {
				locStr = fmt.Sprintf("%s:%d", file, line)
			}
		}
		b.WriteString(fmt.Sprintf("  --> %s\n\n", colorize("\033[36m", locStr)))
	}

	// Context details
	hasContext := false
	if provider != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n", colorize("\033[1m", "Provider:  "), provider))
		hasContext = true
	}
	if dependsOn != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n", colorize("\033[1m", "Dependency:"), dependsOn))
		hasContext = true
	}
	if field != "" && field != dependsOn {
		b.WriteString(fmt.Sprintf("  %s %s\n", colorize("\033[1m", "Field:     "), field))
		hasContext = true
	}

	// Helpful hint
	hint := generateHint(title, message, provider, dependsOn)
	if hint != "" {
		if hasContext {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", colorize("\033[1;33m", "💡 Hint:"), colorize("\033[33m", hint)))
	}

	// Cause chain if more than one error
	if len(chain) > 1 {
		b.WriteString("\n")
		b.WriteString(colorize("\033[90m", "  Caused by:\n"))
		for i := 1; i < len(chain); i++ {
			b.WriteString(fmt.Sprintf("  %s %s\n", colorize("\033[90m", "•"), chain[i].Error()))
		}
	}

	return b.String()
}

func formatGenerateErrorKind(k generator.ErrorKind) string {
	switch k {
	case generator.ErrorDependencyResolve:
		return "Dependency Resolution Error"
	case generator.ErrorValidation:
		return "Validation Error"
	case generator.ErrorGraphBuild:
		return "Graph Build Error"
	case generator.ErrorCodeGeneration:
		return "Code Generation Error"
	default:
		return "Generation Error"
	}
}

func formatParseErrorKind(k parser.ParseErrorKind) string {
	switch k {
	case parser.ParseErrorValidation:
		return "Validation Error"
	case parser.ParseErrorPackageLoad:
		return "Package Load Error"
	case parser.ParseErrorWorkspace:
		return "Workspace Error"
	default:
		return "Parser Error"
	}
}

func parseLocationString(raw string) (path string, line, col int) {
	parts := strings.Split(raw, ":")
	if len(parts) >= 3 {
		c, errC := strconv.Atoi(parts[len(parts)-1])
		l, errL := strconv.Atoi(parts[len(parts)-2])
		if errC == nil && errL == nil {
			return strings.Join(parts[:len(parts)-2], ":"), l, c
		}
	}
	if len(parts) >= 2 {
		l, errL := strconv.Atoi(parts[len(parts)-1])
		if errL == nil {
			return strings.Join(parts[:len(parts)-1], ":"), l, 0
		}
	}
	return raw, 0, 0
}

func renderSourceSnippet(filePath string, targetLine, targetCol int, highlight string) string {
	if filePath == "" || targetLine <= 0 {
		return ""
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		// Try resolving relative path if needed
		if !filepath.IsAbs(filePath) {
			if cwd, cwdErr := os.Getwd(); cwdErr == nil {
				content, err = os.ReadFile(filepath.Join(cwd, filePath))
			}
		}
		if err != nil {
			return ""
		}
	}

	rawLines := strings.Split(string(content), "\n")
	if targetLine > len(rawLines) {
		return ""
	}

	start := targetLine - 1
	if start < 1 {
		start = 1
	}
	end := targetLine + 1
	if end > len(rawLines) {
		end = len(rawLines)
	}

	gutterWidth := len(strconv.Itoa(end))
	if gutterWidth < 2 {
		gutterWidth = 2
	}

	var b strings.Builder
	locStr := filePath
	if targetCol > 0 {
		locStr = fmt.Sprintf("%s:%d:%d", filePath, targetLine, targetCol)
	} else {
		locStr = fmt.Sprintf("%s:%d", filePath, targetLine)
	}

	b.WriteString(fmt.Sprintf("  --> %s\n", colorize("\033[36m", locStr)))
	b.WriteString(fmt.Sprintf("  %s │\n", strings.Repeat(" ", gutterWidth)))

	for l := start; l <= end; l++ {
		lineContent := rawLines[l-1]
		lineContent = strings.TrimRight(lineContent, "\r")

		if l == targetLine {
			b.WriteString(fmt.Sprintf("  %s │ %s\n", colorize("\033[1;31m", fmt.Sprintf("%*d", gutterWidth, l)), lineContent))

			// Draw caret pointer
			pointerCol := targetCol
			if pointerCol <= 0 && highlight != "" {
				idx := strings.Index(lineContent, highlight)
				if idx >= 0 {
					pointerCol = idx + 1
				}
			}

			if pointerCol > 0 {
				pointerLen := 1
				if highlight != "" {
					pointerLen = len(highlight)
				}
				if pointerCol <= len(lineContent) {
					// Count spaces, keeping tabs aligned
					var indent strings.Builder
					for i, ch := range lineContent {
						if i >= pointerCol-1 {
							break
						}
						if ch == '\t' {
							indent.WriteString("\t")
						} else {
							indent.WriteString(" ")
						}
					}
					caret := colorize("\033[1;31m", strings.Repeat("^", pointerLen))
					b.WriteString(fmt.Sprintf("  %s │ %s%s\n", strings.Repeat(" ", gutterWidth), indent.String(), caret))
				}
			}
		} else {
			b.WriteString(fmt.Sprintf("  %s │ %s\n", colorize("\033[90m", fmt.Sprintf("%*d", gutterWidth, l)), lineContent))
		}
	}

	b.WriteString(fmt.Sprintf("  %s │\n", strings.Repeat(" ", gutterWidth)))
	return b.String()
}

func generateHint(kindStr, message, provider, dependsOn string) string {
	msgLower := strings.ToLower(message)
	if strings.Contains(msgLower, "provider not found for dependency") || strings.Contains(msgLower, "missing dependency") {
		return "Ensure a provider function returning this type is defined, annotated with @Injectable, and its package is not ignored in dix.config.json."
	}
	if strings.Contains(msgLower, "circular dependency") {
		return "A cycle was detected in the dependency graph. Break the cycle by refactoring dependencies or wrapping shared singletons in di.Singleton[T]."
	}
	if strings.Contains(msgLower, "cannot find @root") {
		return "Dix needs an entrypoint. Annotate your top-level composition root function with both @Injectable and @Root."
	}
	if strings.Contains(msgLower, "multiple @root") {
		return "Dix only supports a single @Root entrypoint per application. Remove @Root from duplicate functions."
	}
	if strings.Contains(msgLower, "duplicate provider") {
		return "Dix requires unique providers per type. Remove or rename the conflicting provider, or annotate it with @Disable."
	}
	if strings.Contains(msgLower, "singleton dependency must be `di.singleton[t]`") {
		return "Dix singletons must be accepted by value as di.Singleton[T], not as a pointer *di.Singleton[T]."
	}
	if strings.Contains(msgLower, "must return one value or (value, error)") {
		return "Provider functions must return either (T) or (T, error)."
	}
	if strings.Contains(msgLower, "second provider return value must be error") {
		return "The second return value of a provider function must be the builtin 'error' type."
	}
	if strings.Contains(msgLower, "disabled provider cannot be used") {
		return "This provider is marked with @Disable and cannot be injected. Remove @Disable if it should be used."
	}
	return ""
}

func colorize(code, text string) string {
	if color.NoColor {
		return text
	}
	return code + text + "\033[0m"
}
