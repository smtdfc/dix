package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type ParseErrorKind string

const (
	ParseErrorValidation  ParseErrorKind = "validation"
	ParseErrorPackageLoad ParseErrorKind = "package_load"
	ParseErrorWorkspace   ParseErrorKind = "workspace"
)

type ParseError struct {
	Kind     ParseErrorKind
	Message  string
	Function string
	Field    string
	File     string
	Line     int
	Column   int
	Cause    error
}

func (e *ParseError) LocationString() string {
	if e == nil || e.File == "" {
		return ""
	}
	if e.Line > 0 && !strings.Contains(e.File, ":") {
		if e.Column > 0 {
			return fmt.Sprintf("%s:%d:%d", e.File, e.Line, e.Column)
		}
		return fmt.Sprintf("%s:%d", e.File, e.Line)
	}
	return e.File
}

func (e *ParseError) Error() string {
	location := ""
	if loc := e.LocationString(); loc != "" {
		location = fmt.Sprintf(" (%s)", loc)
	}

	context := ""
	if e.Function != "" {
		context = fmt.Sprintf(" [fn=%s]", e.Function)
	}
	if e.Field != "" {
		context += fmt.Sprintf(" [field=%s]", e.Field)
	}

	msg := fmt.Sprintf("%s%s%s", e.Message, context, location)
	return msg
}

func (e *ParseError) Unwrap() error { return e.Cause }

func NewParseError(kind ParseErrorKind, msg string) *ParseError {
	return &ParseError{
		Kind:    kind,
		Message: msg,
	}
}

func parseFileLocation(loc string) (file string, line, col int) {
	if loc == "" {
		return "", 0, 0
	}
	parts := strings.Split(loc, ":")
	if len(parts) >= 3 {
		c, errC := strconv.Atoi(parts[len(parts)-1])
		l, errL := strconv.Atoi(parts[len(parts)-2])
		if errC == nil && errL == nil {
			filePath := strings.Join(parts[:len(parts)-2], ":")
			return filePath, l, c
		}
	}
	if len(parts) >= 2 {
		l, errL := strconv.Atoi(parts[len(parts)-1])
		if errL == nil {
			filePath := strings.Join(parts[:len(parts)-1], ":")
			return filePath, l, 0
		}
	}
	return loc, 0, 0
}

func NewValidationError(msg, fnName, fieldName, file string) *ParseError {
	_, line, col := parseFileLocation(file)
	return &ParseError{
		Kind:     ParseErrorValidation,
		Message:  msg,
		Function: fnName,
		Field:    fieldName,
		File:     file,
		Line:     line,
		Column:   col,
	}
}

func NewValidationErrorWithLocation(msg, fnName, fieldName, file string, line, column int) *ParseError {
	return &ParseError{
		Kind:     ParseErrorValidation,
		Message:  msg,
		Function: fnName,
		Field:    fieldName,
		File:     file,
		Line:     line,
		Column:   column,
	}
}

func NewPackageLoadError(cause error) *ParseError {
	return &ParseError{
		Kind:    ParseErrorPackageLoad,
		Message: "failed to load Go packages",
		Cause:   cause,
	}
}
