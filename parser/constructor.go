package parser

type Constructor struct {
	File   string        `json:"file"`
	Name   string        `json:"name"`
	Deps   []*Dependency `json:"deps"`
	Return *ReturnValue  `json:"return"`
}
