package parser

type Constructor struct {
	File   string
	Name   string
	Deps   []*Dependency
	Return *ReturnValue
}
