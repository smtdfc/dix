package dix

type AnnotationMetadata struct {
	Key   string
	Value string
	File  string
	Line  int
	Path  string
	Pos   string
}

type Annotation interface {
	Type() string
}

type WireAnnotation struct {
	Path   string
	Target string
	Deps   []string
	File   string
	Pos    string
}

// Type returns the annotation type.
func (w *WireAnnotation) Type() string {
	return "Wire"
}

type FactoryAnnotation struct {
	Path     string
	Function string
	Alias    string
	File     string
	Pos      string
}

// Type returns the annotation type.
func (f *FactoryAnnotation) Type() string {
	return "Factory"
}

type FinalAnnotation struct {
	Path   string
	Target string
	File   string
	Pos    string
}

// Type returns the annotation type.
func (w *FinalAnnotation) Type() string {
	return "Final"
}

type DisableAnnotation struct {
	Path   string
	Target string
	File   string
	Pos    string
}

// Type returns the annotation type.
func (w *DisableAnnotation) Type() string {
	return "Disable"
}
