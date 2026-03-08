package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/smtdfc/dix/generator"
	"github.com/smtdfc/dix/parser"
)

func fatalDixError(err error) {
	if err == nil {
		return
	}

	var parseErr *parser.ParseError
	if errors.As(err, &parseErr) {
		fmt.Fprintf(os.Stderr, "\033[31m[Error]\033[0m %s\n", parseErr.Error())
		os.Exit(1)
	}

	var genErr *generator.GenerateError
	if errors.As(err, &genErr) {
		fmt.Fprintf(os.Stderr, "\033[31m[Error]\033[0m %s\n", genErr.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\033[31m[Error]\033[0m unknown: %v\n", err)
	os.Exit(1)
}
