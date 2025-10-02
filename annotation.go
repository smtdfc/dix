package dix

type AnnotationMetadata struct {
	Key   string
	Value string
	File  string
	Line  int
	Path  string
}

type Annotation interface {
	Type() string
}

type WireAnnotation struct {
	Path   string
	Target string
	Deps   []string
}

func (w *WireAnnotation) Type() string {
	return "Wire"
}

type FactoryAnnotation struct {
	Path     string
	Function string
	Alias    string
}

func (f *FactoryAnnotation) Type() string {
	return "Factory"
}
