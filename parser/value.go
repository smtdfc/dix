package parser

import "fmt"

type Dependency struct {
	Name        string    `json:"name"`
	Type        *TypeInfo `json:"type"`
	IsSingleton bool      `json:"is_sng"`
}

func (d *Dependency) String() string {
	return fmt.Sprintf(" %s (%s)", d.Name, d.Type.String())
}

type ReturnValue struct {
	Type *TypeInfo `json:"type"`
}
