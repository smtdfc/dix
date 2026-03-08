package di

type Singleton[T any] struct {
	Instance T
}

func NewSingleton[T any](instance T) Singleton[T] {
	return Singleton[T]{
		Instance: instance,
	}
}

func (s Singleton[T]) Get() T {
	return s.Instance
}
