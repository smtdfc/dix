package parser

import "fmt"

type Provider struct {
	File         string        `json:"file"`
	Line         int           `json:"line,omitempty"`
	Column       int           `json:"column,omitempty"`
	Name         string        `json:"name"`
	Deps         []*Dependency `json:"deps"`
	Return       *ReturnValue  `json:"return"`
	PackagePath  string        `json:"pkg_path"`
	PackageName  string        `json:"pkg_name"`
	IsDisable    bool          `json:"is_disable"`
	ReturnsError bool          `json:"returns_error"`
}

func (p *Provider) LocationString() string {
	if p == nil || p.File == "" {
		return ""
	}
	if p.Line > 0 {
		if p.Column > 0 {
			return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
		}
		return fmt.Sprintf("%s:%d", p.File, p.Line)
	}
	return p.File
}

