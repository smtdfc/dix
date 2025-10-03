package dix

type Dependency struct {
	Name       string
	Standalone bool
}

type Factory struct {
	Alias    string
	Function string
	Deps     []*Dependency
	Module   string
	Final    bool
	Disable  bool
	File     string
	Pos      string
}

type DIConfig struct {
	Container map[string]*Factory
}
