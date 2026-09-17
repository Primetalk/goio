package resource

import "github.com/primetalk/goio/io"

// Map transforms the acquired value while preserving its release action.
func (res Resource[A]) Map[B any](f func(A) B) Resource[B] {
	return Map(res, f)
}

// FlatMap composes resources and preserves reverse-order release semantics.
func (res Resource[A]) FlatMap[B any](f func(A) Resource[B]) Resource[B] {
	return FlatMap(res, f)
}

// Use runs f with the acquired value and releases the resource afterward.
func (res Resource[A]) Use[B any](f func(A) io.IO[B]) io.IO[B] {
	return Use(res, f)
}
