package test

type A struct{}

// @Injectable
func NewA() *A {
	return &A{}
}
