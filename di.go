package dix

type Dependency struct {
	Name       string
	Standalone bool
}

type Factory struct {
	Function string
	Deps     []*Dependency
	Module   string
}

type DIConfig struct {
	Container map[string]*Factory
}
