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
