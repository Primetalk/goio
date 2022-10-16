package stream

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
)

// Empty returns an empty stream.
func Empty[A any]() Stream[A] {
	return Stream[A](io.Pure(func() StepResult[A] { return NewStepResultFinished[A]() }))
}

var empty = Empty[fun.Unit]()

// EmptyUnit returns an empty stream of units.
// It's more performant because the same instance is being used.
func EmptyUnit() Stream[fun.Unit] {
	return empty
}

// Eval returns a stream of one value that is the result of IO.
func Eval[A any](ioa io.IO[A]) Stream[A] {
	return Stream[A](io.Map(ioa, func(a A) StepResult[A] {
		return NewStepResult(a, Empty[A]())
	}))
}

// Lift returns a stream of one value.
func Lift[A any](a A) Stream[A] {
	return Emit(a)
}

// LiftMany returns a stream with all the given values.
func LiftMany[A any](as ...A) Stream[A] {
	return EmitMany(as...)
}

// Emit returns a stream of a single element
func Emit[A any](a A) Stream[A] {
	return Eval(io.Lift(a))
}

// EmitMany returns a stream with all the given values.
func EmitMany[A any](as ...A) Stream[A] {
	return FromSlice(as)
}

// FromSlice constructs a stream from the slice.
func FromSlice[A any](as []A) Stream[A] {
	if len(as) == 0 {
		return Empty[A]()
	} else {
		return Stream[A](io.Lift(NewStepResult(as[0], FromSlice(as[1:]))))
	}
}

// Generate constructs an infinite stream of values using the production function.
func Generate[A any, S any](zero S, f func(s S) (S, A)) Stream[A] {
	return Stream[A](io.Eval(func() (StepResult[A], error) {
		s, a := f(zero)
		return NewStepResult(a, Generate(s, f)), nil
	}))
}

// Unfold constructs an infinite stream of values using the production function.
func Unfold[A any](zero A, f func(A) A) Stream[A] {
	return Generate(zero, func(s A) (A, A) {
		r := f(s)
		return r, r
	})
}

// FromSideEffectfulFunction constructs a stream from a Go-style function.
// It is expected that this function is not pure and can return different results.
func FromSideEffectfulFunction[A any](f func() (A, error)) Stream[A] {
	return Repeat(Eval(io.Eval(f)))
}

// FromStepResult constructs a stream from an IO that returns StepResult.
func FromStepResult[A any](iosr io.IO[StepResult[A]]) Stream[A] {
	return Stream[A](iosr)
}

// Nats returns an infinite stream of ints starting from 1.
func Nats() Stream[int] {
	return Unfold(0, func(s int) int {
		return s + 1
	})
}
