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
