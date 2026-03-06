package parser

import "errors"

type ParseError struct {
	error
}

func NewParseError(msg string) *ParseError {
	return &ParseError{
		error: errors.New(msg),
	}
}
