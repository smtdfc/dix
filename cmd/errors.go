package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/smtdfc/dix/generator"
	"github.com/smtdfc/dix/parser"
)

func fatalDixError(err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "\033[31m[Error]\033[0m\n%s\n", formatDixError(err))
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
