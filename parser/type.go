package parser

type TypeInfo struct {
	Name      string `json:"name"`
	Pkg       string `json:"pkg"`
	IsPointer bool   `json:"is_ptr"`
}
