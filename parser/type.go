package parser

import "fmt"

type TypeInfo struct {
	Name      string `json:"name"`
	Pkg       string `json:"pkg"`
	IsPointer bool   `json:"is_ptr"`
}

func (t *TypeInfo) Signature() string {
	if t.IsPointer {
		return fmt.Sprintf("ptr_%s@%s", t.Name, t.Pkg)
	}
	return fmt.Sprintf("%s@%s", t.Name, t.Pkg)
}

func (t *TypeInfo) String() string {
	if t.IsPointer {
		return fmt.Sprintf("type of *%s in package %s", t.Name, t.Pkg)
	}
	return fmt.Sprintf("type of %s in package %s", t.Name, t.Pkg)
}
