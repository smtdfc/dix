package parser

import "fmt"

type Dependency struct {
	Name        string    `json:"name"`
	Type        *TypeInfo `json:"type"`
	IsSingleton bool      `json:"is_sng"`
	File        string    `json:"file,omitempty"`
	Line        int       `json:"line,omitempty"`
	Column      int       `json:"column,omitempty"`
}

func (d *Dependency) String() string {
	return fmt.Sprintf(" %s (%s)", d.Name, d.Type.String())
}

func (d *Dependency) LocationString() string {
	if d == nil || d.File == "" {
		return ""
	}
	if d.Line > 0 {
		if d.Column > 0 {
			return fmt.Sprintf("%s:%d:%d", d.File, d.Line, d.Column)
		}
		return fmt.Sprintf("%s:%d", d.File, d.Line)
	}
	return d.File
}

type ReturnValue struct {
	Type *TypeInfo `json:"type"`
}
