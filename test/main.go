package test

type A struct{}

// @Injectable
// @Root
func NewA() *A {
	return &A{}
}
