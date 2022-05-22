// Package stream provides a way to construct data processing streams
// from smaller pieces.
// The design is inspired by fs2 Scala library.
package stream

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
)

// Stream is modelled as a function that performs a single step in the state machine.
type Stream[A any] io.IO[StepResult[A]]

// StepResult[A] represents the result of a single step in the step machine.
// It might be one of - empty, new value, or finished.
// The step result also returns the continuation of the stream.
type StepResult[A any] struct {
	Value        A
	Continuation Stream[A]
	HasValue     bool
	IsFinished   bool
}

// NewStepResult constructs StepResult that has one value.
func NewStepResult[A any](value A, continuation Stream[A]) StepResult[A] {
	return StepResult[A]{
		Value:        value,
		HasValue:     true,
		Continuation: continuation,
		IsFinished:   false,
	}
}

// NewStepResultEmpty constructs an empty StepResult.
func NewStepResultEmpty[A any](continuation Stream[A]) StepResult[A] {
	return StepResult[A]{
		HasValue:     false,
		Continuation: continuation,
		IsFinished:   false,
	}
}

// NewStepResultFinished constructs a finished StepResult.
// The continuation will be empty as well.
func NewStepResultFinished[A any]() StepResult[A] {
	return StepResult[A]{
		IsFinished:   true,
		Continuation: Empty[A](),
	}
}

// MapEval maps the values of the stream. The provided function returns an IO.
func MapEval[A any, B any](stm Stream[A], f func(a A) io.IO[B]) Stream[B] {
	return io.FlatMap[StepResult[A]](
		stm,
		func(sra StepResult[A]) io.IO[StepResult[B]] {
			if sra.IsFinished {
				return io.Lift(NewStepResultFinished[B]())
			} else if sra.HasValue {
				iob := f(sra.Value)
				return io.Map(iob, func(b B) StepResult[B] {
					return NewStepResult(b, MapEval(sra.Continuation, f))
				})
			} else {
				return io.Lift(
					NewStepResultEmpty(MapEval(sra.Continuation, f)),
				)
			}
		})
}

// Map converts values of the stream.
func Map[A any, B any](stm Stream[A], f func(a A) B) Stream[B] {
	return MapEval(stm, func(a A) io.IO[B] { return io.Lift(f(a)) })
}

// AndThen appends another stream after the end of the first one.
func AndThen[A any](stm1 Stream[A], stm2 Stream[A]) Stream[A] {
	return AndThenLazy(stm1, func() Stream[A] { return stm2 })
}

// AndThenLazy appends another stream. The other stream is constructed lazily.
func AndThenLazy[A any](stm1 Stream[A], stm2 func() Stream[A]) Stream[A] {
	return io.FlatMap[StepResult[A]](
		stm1,
		func(sra StepResult[A]) io.IO[StepResult[A]] {
			if sra.IsFinished {
				return stm2()
			} else {
				return io.Lift(StepResult[A]{
					Value:        sra.Value,
					Continuation: AndThenLazy(sra.Continuation, stm2),
					HasValue:     sra.HasValue,
				})
			}
		})

}

// FlatMap constructs a
func FlatMap[A any, B any](stm Stream[A], f func(a A) Stream[B]) Stream[B] {
	return io.FlatMap[StepResult[A]](
		stm,
		func(sra StepResult[A]) io.IO[StepResult[B]] {
			if sra.IsFinished {
				return io.Lift(NewStepResultFinished[B]())
			} else if sra.HasValue {
				stmb1 := f(sra.Value)
				stmb := AndThenLazy(stmb1, func() Stream[B] { return FlatMap(sra.Continuation, f) })
				return stmb
			} else {
				return io.Lift(NewStepResultEmpty(FlatMap(sra.Continuation, f)))
			}
		})
}

// Flatten simplifies a stream of streams to just the stream of values by concatenating all
// inner streams.
func Flatten[A any](stm Stream[Stream[A]]) Stream[A] {
	return FlatMap(stm, fun.Identity[Stream[A]])
}

// StateFlatMap maintains state along the way.
func StateFlatMap[A any, B any, S any](stm Stream[A], zero S, f func(a A, s S) (S, Stream[B])) Stream[B] {
	return StateFlatMapWithFinish(stm, zero, f, func(S) Stream[B] { return Empty[B]() })
}

// StateFlatMapWithFinish maintains state along the way.
// When the source stream finishes, it invokes onFinish with the last state.
func StateFlatMapWithFinish[A any, B any, S any](stm Stream[A], zero S, f func(a A, s S) (S, Stream[B]), onFinish func(s S) Stream[B]) Stream[B] {
	res := io.FlatMap[StepResult[A]](
		stm,
		func(sra StepResult[A]) (iores io.IO[StepResult[B]]) {
			if sra.IsFinished {
				iores = io.Lift(NewStepResultEmpty(onFinish(zero)))
			} else if sra.HasValue {
				st, stmb1 := f(sra.Value, zero)
				stmb := AndThenLazy(stmb1, func() Stream[B] { return StateFlatMapWithFinish(sra.Continuation, st, f, onFinish) })
				iores = stmb
			} else {
				iores = io.Lift(NewStepResultEmpty(StateFlatMap(sra.Continuation, zero, f)))
			}
			return
		})
	return res.(Stream[B])
}

// Filter leaves in the stream only the elements that satisfy the given predicate.
func Filter[A any](stm Stream[A], predicate func(A) bool) Stream[A] {
	return io.Map[StepResult[A]](
		stm,
		func(sra StepResult[A]) StepResult[A] {
			if sra.IsFinished {
				return NewStepResultFinished[A]()
			} else {
				cont := Filter(sra.Continuation, predicate)
				if sra.HasValue && predicate(sra.Value) {
					return NewStepResult(sra.Value, cont)
				} else {
					return NewStepResultEmpty(cont)
				}
			}
		})
}
