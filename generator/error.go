package generator

import "fmt"

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
	Cause     error
}

func (e *GenerateError) Error() string {
	context := ""
	if e.Provider != "" {
		context += fmt.Sprintf(" [provider=%s]", e.Provider)
	}
	if e.DependsOn != "" {
		context += fmt.Sprintf(" [depends_on=%s]", e.DependsOn)
	}

	msg := fmt.Sprintf("generator/%s: %s%s", e.Kind, e.Message, context)
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
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
