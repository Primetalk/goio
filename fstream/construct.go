package fstream

import "github.com/primetalk/goio/stream"

// Nats returns an infinite stream of ints starting from 1.
func Nats() Stream[int] {
	return LiftStream(stream.Nats())
}

// Generate constructs an infinite stream of values using the production function.
func Generate[A any, S any](zero S, f func(s S) (S, A)) Stream[A] {
	return LiftStream(stream.Generate(zero, f))
}

// Unfold constructs an infinite stream of values using the production function.
func Unfold[A any](zero A, f func(A) A) Stream[A] {
	return LiftStream(stream.Unfold(zero, f))
}
