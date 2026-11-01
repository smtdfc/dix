package generator

import (
	"fmt"
	"strings"
)

type ErrorKind string

const (
	ErrorValidation        ErrorKind = "validation"
	ErrorGraphBuild        ErrorKind = "graph_build"
	ErrorDependencyResolve ErrorKind = "dependency_resolution"
	ErrorCodeGeneration    ErrorKind = "code_generation"
)

type GenerateError struct {
	Kind      ErrorKind
	Message   string
	Provider  string
	DependsOn string
	File      string
	Line      int
	Column    int
	Cause     error
}

func (e *GenerateError) LocationString() string {
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

func (e *GenerateError) Error() string {
	context := ""
	if e.Provider != "" {
		context += fmt.Sprintf(" [provider=%s]", e.Provider)
	}
	if e.DependsOn != "" {
		context += fmt.Sprintf(" [depends_on=%s]", e.DependsOn)
	}
	location := ""
	if loc := e.LocationString(); loc != "" {
		location = fmt.Sprintf(" (%s)", loc)
	}

	msg := fmt.Sprintf("generator/%s: %s%s%s", e.Kind, e.Message, context, location)
	return msg
}

func (e *GenerateError) Unwrap() error { return e.Cause }

func (e *GenerateError) WithLocation(file string, line, column int) *GenerateError {
	e.File = file
	e.Line = line
	e.Column = column
	return e
}

func NewGenerateError(kind ErrorKind, msg, provider, dependsOn string, cause error) *GenerateError {
	return &GenerateError{
		Kind:      kind,
		Message:   msg,
		Provider:  provider,
		DependsOn: dependsOn,
		Cause:     cause,
	}
}

func NewGenerateErrorWithLocation(kind ErrorKind, msg, provider, dependsOn, file string, line, column int, cause error) *GenerateError {
	return &GenerateError{
		Kind:      kind,
		Message:   msg,
		Provider:  provider,
		DependsOn: dependsOn,
		File:      file,
		Line:      line,
		Column:    column,
		Cause:     cause,
	}
}

