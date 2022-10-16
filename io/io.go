// Package io implements IO tools similar to what is available in Scala cats library (and Haskell IO).
package io

import (
	"log"

	"github.com/pkg/errors"
	"github.com/primetalk/goio/fun"
)

// IO[A] represents a calculation that will yield a value of type A once executed.
// The calculation might as well fail.
// It is designed to not panic ever.
type IO[A any] Continuation[A]

// LiftPair[A] constructs an IO from constant values.
func LiftPair[A any](a A, err error) IO[A] {
	return func() ResultOrContinuation[A] {
		return ResultOrContinuation[A]{
			Value:        a,
			Error:        err,
			Continuation: nil,
		}
	}
}

// UnsafeRunSync runs the given IO[A] synchronously and returns the result.
func UnsafeRunSync[A any](io IO[A]) (res A, err error) {
	defer fun.RecoverToErrorVar("UnsafeRunSync", &err)
	return ObtainResult(Continuation[A](io))
}

// RunSync is the same as UnsafeRunSync but returns GoResult[A].
func RunSync[A any](io IO[A]) GoResult[A] {
	a, err := UnsafeRunSync(io)
	return GoResult[A]{Value: a, Error: err}
}

// Delay[A] wraps a function that will then return an IO.
func Delay[A any](f func() IO[A]) IO[A] {
	return func() ResultOrContinuation[A] {
		return f()()
	}
}

// Eval[A] constructs an IO[A] from a simple function that might fail.
// If there is panic in the function, it's recovered from and represented as an error.
func Eval[A any](f func() (A, error)) IO[A] {
	return func() ResultOrContinuation[A] {
		a, err := f()
		return ResultOrContinuation[A]{
			Value: a,
			Error: err,
		}
	}
}

// FromPureEffect constructs IO from the simplest function signature.
func FromPureEffect(f func()) IO[fun.Unit] {
	return func() ResultOrContinuation[fun.Unit] {
		f()
		return ResultOrContinuation[fun.Unit]{}
	}
}

// FromUnit consturcts IO[fun.Unit] from a simple function that might fail.
func FromUnit(f func() error) IO[fun.Unit] {
	return func() ResultOrContinuation[fun.Unit] {
		return ResultOrContinuation[fun.Unit]{
			Error: f(),
		}
	}
}

// Pure[A] constructs an IO[A] from a function that cannot fail.
func Pure[A any](f func() A) IO[A] {
	return Eval(func() (A, error) {
		return f(), nil
	})
}

// FromConstantGoResult converts an existing GoResult value into a fake IO.
// NB! This is not for normal delayed IO execution!
func FromConstantGoResult[A any](gr GoResult[A]) IO[A] {
	return Eval(func() (A, error) { return gr.Value, gr.Error })
}

// MapErr maps the result of IO[A] using a function that might fail.
func MapErr[A any, B any](ioA IO[A], f func(a A) (B, error)) IO[B] {
	return func() ResultOrContinuation[B] {
		a, err := ObtainResult(Continuation[A](ioA))
		if err == nil {
			cont := Continuation[B](func() ResultOrContinuation[B] {
				b, err := f(a)
				return ResultOrContinuation[B]{
					Value: b,
					Error: err,
				}
			})
			return ResultOrContinuation[B]{
				Continuation: &cont,
			}
		} else {
			return ResultOrContinuation[B]{
				Error: err,
			}
		}
	}
}

// Map converts the IO[A] result using the provided function that cannot fail.
func Map[A any, B any](ioA IO[A], f func(a A) B) IO[B] {
	return MapErr(ioA, func(a A) (B, error) { return f(a), nil })
}

// MapConst ignores the result and replaces it with the given constant.
func MapConst[A any, B any](ioA IO[A], b B) IO[B] {
	return Map(ioA, fun.Const[A](b))
}

// FlatMap converts the result of IO[A] using a function that itself returns an IO[B].
// It'll fail if any of IO[A] or IO[B] fail.
func FlatMap[A any, B any](ioA IO[A], f func(a A) IO[B]) IO[B] {
	return func() ResultOrContinuation[B] {
		a, err := ObtainResult(Continuation[A](ioA))
		if err == nil {
			cont := Continuation[B](func() ResultOrContinuation[B] {
				ioB := f(a)
				return ioB()
			})
			return ResultOrContinuation[B]{
				Continuation: &cont,
			}
		} else {
			return ResultOrContinuation[B]{
				Error: err,
			}
		}
	}
}

// FlatMapErr converts IO[A] result using a function that might fail.
// It seems to be identical to MapErr.
func FlatMapErr[A any, B any](ioA IO[A], f func(a A) (B, error)) IO[B] {
	return FlatMap(ioA, func(a A) IO[B] { return LiftPair(f(a)) })
}

// AndThen runs the first IO, ignores it's result and then runs the second one.
func AndThen[A any, B any](ioa IO[A], iob IO[B]) IO[B] {
	return FlatMap(ioa, func(A) IO[B] {
		return iob
	})
}

// Lift[A] constructs an IO[A] from a constant value.
func Lift[A any](a A) IO[A] {
	return LiftPair(a, nil)
}

// LiftFunc wraps the result of function into IO.
func LiftFunc[A any, B any](f func(A) B) func(A) IO[B] {
	return func(a A) IO[B] {
		return Lift(f(a))
	}
}

// Fail[A] constructs an IO[A] that fails with the given error.
func Fail[A any](err error) IO[A] {
	var a A
	return LiftPair(a, err)
}

// Fold performs different calculations based on whether IO[A] failed or succeeded.
func Fold[A any, B any](ioA IO[A], f func(a A) IO[B], recover func(error) IO[B]) IO[B] {
	return func() ResultOrContinuation[B] {
		a, err := ObtainResult(Continuation[A](ioA))
		var cont Continuation[B]
		if err == nil {
			cont = Continuation[B](func() ResultOrContinuation[B] {
				ioB := f(a)
				return ioB()
			})
		} else {
			cont = Continuation[B](func() ResultOrContinuation[B] {
				ioB := recover(err)
				return ioB()
			})
		}
		return ResultOrContinuation[B]{
			Continuation: &cont,
		}
	}
}

// FoldErr folds IO using simple Go-style functions that might fail.
func FoldErr[A any, B any](ioA IO[A], f func(a A) (B, error), recover func(error) (B, error)) IO[B] {
	return Fold(ioA,
		func(a A) IO[B] { return LiftPair(f(a)) },
		func(err error) IO[B] { return LiftPair(recover(err)) },
	)
}

// Sequence takes a slice of IOs and returns an IO that will contain a slice of results.
// It'll fail if any of the internal computations fail.
func Sequence[A any](ioas []IO[A]) (res IO[[]A]) {
	res = Lift([]A{})
	for _, ioa := range ioas {
		ioaCopy := ioa // See https://eli.thegreenplace.net/2019/go-internals-capturing-loop-variables-in-closures/
		res = FlatMap(res, func(as []A) IO[[]A] {
			return Map(ioaCopy, func(a A) []A {
				return append(as, a)
			})
		})
	}
	return
}

// SequenceUnit takes a slice of IO units and returns IO that executes all of them.
// It'll fail if any of the internal computations fail.
func SequenceUnit(ious []IOUnit) (res IOUnit) {
	res = IOUnit1
	for _, iou := range ious {
		iou1 := iou // See https://eli.thegreenplace.net/2019/go-internals-capturing-loop-variables-in-closures/
		res = FlatMap(res, func(fun.Unit) IOUnit { return iou1 })
	}
	return
}

var ErrorNPE = errors.New("nil pointer")

// Unptr retrieves the value at pointer. Fails if nil
func Unptr[A any](ptra *A) IO[A] {
	if ptra == nil {
		return Fail[A](ErrorNPE)
	} else {
		return Lift(*ptra)
	}
}

// Wrapf wraps an error with additional context information
func Wrapf[A any](io IO[A], format string, args ...interface{}) IO[A] {
	return Recover(io, func(err error) IO[A] {
		return Fail[A](errors.Wrapf(err, format, args...))
	})
}

// IOUnit1 is a IO[Unit] that will always return Unit1.
var IOUnit1 = Lift(fun.Unit1)

// IOUnit is IO[Unit]
type IOUnit = IO[fun.Unit]

// ForEach calls the provided callback after IO is completed.
func ForEach[A any](io IO[A], cb func(a A)) IO[fun.Unit] {
	return Map(io, func(a A) fun.Unit {
		cb(a)
		return fun.Unit1
	})
}

// Finally runs the finalizer regardless of the success of the IO.
// In case finalizer fails as well, the second error is printed to log.
func Finally[A any](io IO[A], finalizer IO[fun.Unit]) IO[A] {
	return Fold(io,
		func(a A) IO[A] {
			return Map(finalizer, fun.ConstUnit(a))
		},
		func(err error) IO[A] {
			return Fold(finalizer,
				func(fun.Unit) IO[A] {
					return Fail[A](err)
				},
				func(err2 error) IO[A] {
					log.Printf("double error during Finally: %+v", err2)
					return Fail[A](err)
				})

		})
}

// Ignore throws away the result of IO.
func Ignore[A any](ioa IO[A]) IOUnit {
	return Map(ioa, fun.Const[A](fun.Unit1))
}
