package parser

type Provider struct {
	File        string        `json:"file"`
	Name        string        `json:"name"`
	Deps        []*Dependency `json:"deps"`
	Return      *ReturnValue  `json:"return"`
	PackagePath string        `json:"pkg_path"`
	PackageName string        `json:"pkg_name"`
	IsDisable   bool          `json:"is_disable"`
}
