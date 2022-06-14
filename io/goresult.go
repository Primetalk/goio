package io

// GoResult[A] is a data structure that represents the Go-style result of a function that
// could fail.
type GoResult[A any] struct {
	Value A
	Error error
}

func (e GoResult[A]) unsafeRun() (res A, err error) {
	defer RecoverToErrorVar("GoResult.unsafeRun", &err)
	return e.Value, e.Error
}
