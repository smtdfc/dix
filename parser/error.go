package parser

import "fmt"

type ParseErrorKind string

const (
	ParseErrorValidation  ParseErrorKind = "validation"
	ParseErrorPackageLoad ParseErrorKind = "package_load"
)

type ParseError struct {
	Kind     ParseErrorKind
	Message  string
	Function string
	Field    string
	File     string
	Cause    error
}

func (e *ParseError) Error() string {
	location := ""
	if e.File != "" {
		location = fmt.Sprintf(" (%s)", e.File)
	}

	context := ""
	if e.Function != "" {
		context = fmt.Sprintf(" [fn=%s]", e.Function)
	}
	if e.Field != "" {
		context += fmt.Sprintf(" [field=%s]", e.Field)
	}

	msg := fmt.Sprintf("parser/%s: %s%s%s", e.Kind, e.Message, context, location)
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", msg, e.Cause)
	}

	return msg
}

func NewParseError(kind ParseErrorKind, msg string) *ParseError {
	return &ParseError{
		Kind:    kind,
		Message: msg,
	}
}

func NewValidationError(msg, fnName, fieldName, file string) *ParseError {
	return &ParseError{
		Kind:     ParseErrorValidation,
		Message:  msg,
		Function: fnName,
		Field:    fieldName,
		File:     file,
	}
}

func NewPackageLoadError(cause error) *ParseError {
	return &ParseError{
		Kind:    ParseErrorPackageLoad,
		Message: "failed to load Go packages",
		Cause:   cause,
	}
}
