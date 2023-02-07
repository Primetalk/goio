package io

// GoResult[A] is a data structure that represents the Go-style result of a function that
// could fail.
type GoResult[A any] struct {
	Value A
	Error error
}

// NewGoResult constructs a GoResult.
func NewGoResult[A any](value A) GoResult[A] {
	return GoResult[A]{
		Value: value,
	}
}

// NewFailedGoResult constructs a GoResult with an error.
func NewFailedGoResult[A any](err error) GoResult[A] {
	return GoResult[A]{
		Error: err,
	}
}

// RunSync is the same as UnsafeRunSync but returns GoResult[A].
func RunSync[A any](io IO[A]) GoResult[A] {
	a, err := UnsafeRunSync(io)
	return GoResult[A]{Value: a, Error: err}
}

// FromConstantGoResult converts an existing GoResult value into a fake IO.
// NB! This is not for normal delayed IO execution!
func FromConstantGoResult[A any](gr GoResult[A]) IO[A] {
	return Eval(func() (A, error) { return gr.Value, gr.Error })
}

// IOFuncToGoResult converts a function that returns IO
// to a function that will return GoResult.
func IOFuncToGoResult[A any, B any](f func(a A) IO[B]) func(A) GoResult[B] {
	return func(a A) GoResult[B] {
		return RunSync(f(a))
	}
}
