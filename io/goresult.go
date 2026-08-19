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

// MakeGoResult constructs a GoResult from a value and an error.
// If there's an error, the value is ignored.
func MakeGoResult[A any](value A, err error) GoResult[A] {
	if err != nil {
		return NewFailedGoResult[A](err)
	}
	return NewGoResult(value)
}

// RunSync executes io through the UnsafeRunSync panic-recovering boundary and
// returns its value or error as GoResult[A].
func RunSync[A any](io IO[A]) GoResult[A] {
	return MakeGoResult(UnsafeRunSync(io))
}

// FromConstantGoResult converts an existing GoResult value into a fake IO.
// NB! This is not for normal delayed IO execution!
func FromConstantGoResult[A any](gr GoResult[A]) IO[A] {
	return Eval(func() (A, error) { return gr.Value, gr.Error })
}

// IOFuncToGoResult converts a function that returns IO to a function that runs
// that IO through RunSync and returns GoResult. Calling the returned function
// invokes f before the RunSync boundary; a panic from f itself is therefore not
// recovered by that boundary.
func IOFuncToGoResult[A any, B any](f func(a A) IO[B]) func(A) GoResult[B] {
	return func(a A) GoResult[B] {
		return RunSync(f(a))
	}
}
