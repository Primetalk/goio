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
type IO[A any] interface {
	unsafeRun() (A, error)
}

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

// LiftPair[A] constructs an IO from constant values.
func LiftPair[A any](a A, err error) IO[A] {
	return GoResult[A]{
		Value: a,
		Error: err,
	}
}

// UnsafeRunSync runs the given IO[A] synchronously and returns the result.
func UnsafeRunSync[A any](io IO[A]) (res A, err error) {
	defer RecoverToErrorVar("UnsafeRunSync", &err)
	return io.unsafeRun()
}

// RunSync is the same as UnsafeRunSync but returns GoResult[A].
func RunSync[A any](io IO[A]) GoResult[A] {
	a, err := UnsafeRunSync(io)
	return GoResult[A]{Value: a, Error: err}
}

// Delay[A] wraps a function that will then return an IO.
func Delay[A any](f func() IO[A]) IO[A] {
	return delayImpl[A]{
		f: f,
	}
}

type delayImpl[A any] struct {
	f func() IO[A]
}

func (i delayImpl[A]) unsafeRun() (a A, err error) {
	defer RecoverToErrorVar("Delay.unsafeRun", &err)
	return i.f().unsafeRun()
}

// Eval[A] constructs an IO[A] from a simple function that might fail.
// If there is panic in the function, it's recovered from and represented as an error.
func Eval[A any](f func() (A, error)) IO[A] {
	return evalImpl[A]{
		f: f,
	}
}

type evalImpl[A any] struct {
	f func() (A, error)
}

func (e evalImpl[A]) unsafeRun() (res A, err error) {
	defer RecoverToErrorVar("Eval.unsafeRun", &err)
	return e.f()
}

// FromPureEffect constructs IO from the simplest function signature.
func FromPureEffect(f func()) IO[fun.Unit] {
	return Eval(func() (fun.Unit, error) {
		f()
		return fun.Unit1, nil
	})
}

// FromUnit consturcts IO[fun.Unit] from a simple function that might fail.
func FromUnit(f func() error) IO[fun.Unit] {
	return Eval(func() (fun.Unit, error) {
		return fun.Unit1, f()
	})
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
	return mapImpl[A, B]{
		io: ioA,
		f:  f,
	}
}

// Map converts the IO[A] result using the provided function that cannot fail.
func Map[A any, B any](ioA IO[A], f func(a A) B) IO[B] {
	return mapImpl[A, B]{
		io: ioA,
		f:  func(a A) (B, error) { return f(a), nil },
	}
}

type mapImpl[A any, B any] struct {
	io IO[A]
	f  func(a A) (B, error)
}

func (e mapImpl[A, B]) unsafeRun() (res B, err error) {
	defer RecoverToErrorVar("Map.unsafeRun", &err)
	var a A
	a, err = e.io.unsafeRun()
	if err == nil {
		res, err = e.f(a)
	}
	return
}

// FlatMap converts the result of IO[A] using a function that itself returns an IO[B].
// It'll fail if any of IO[A] or IO[B] fail.
func FlatMap[A any, B any](ioA IO[A], f func(a A) IO[B]) IO[B] {
	return flatMapImpl[A, B]{
		io: ioA,
		f:  f,
	}
}

type flatMapImpl[A any, B any] struct {
	io IO[A]
	f  func(a A) IO[B]
}

func (e flatMapImpl[A, B]) unsafeRun() (res B, err error) {
	defer RecoverToErrorVar("FlatMap.unsafeRun", &err)
	var a A
	a, err = e.io.unsafeRun()
	if err == nil {
		res, err = e.f(a).unsafeRun()
	}
	return
}

// FlatMapErr converts IO[A] result using a function that might fail.
// It seems to be identical to MapErr.
func FlatMapErr[A any, B any](ioA IO[A], f func(a A) (B, error)) IO[B] {
	return flatMapImpl[A, B]{
		io: ioA,
		f:  func(a A) IO[B] { return LiftPair(f(a)) },
	}
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

// Fail[A] constructs an IO[A] that fails with the given error.
func Fail[A any](err error) IO[A] {
	var a A
	return LiftPair(a, err)
}

// Fold performs different calculations based on whether IO[A] failed or succeeded.
func Fold[A any, B any](io IO[A], f func(a A) IO[B], recover func(error) IO[B]) IO[B] {
	return Delay(func() IO[B] {
		var a A
		var err error
		a, err = io.unsafeRun()
		if err == nil {
			return f(a)
		} else {
			return recover(err)
		}
	})
}

// FoldToGoResult converts either value or error to go result
// typically it should never fail.
func FoldToGoResult[A any](io IO[A]) IO[GoResult[A]] {
	return Fold(
		io,
		func(a A) IO[GoResult[A]] {
			return Lift(GoResult[A]{Value: a})
		},
		func(err error) IO[GoResult[A]] {
			return Lift(GoResult[A]{Error: err})
		},
	)
}

// UnfoldGoResult represents GoResult back to ordinary IO.
func UnfoldGoResult[A any](iogr IO[GoResult[A]]) IO[A] {
	return MapErr(iogr, func(gr GoResult[A]) (A, error) { return gr.Value, gr.Error })
}

// FoldErr folds IO using simple Go-style functions that might fail.
func FoldErr[A any, B any](io IO[A], f func(a A) (B, error), recover func(error) (B, error)) IO[B] {
	return Eval(func() (b B, err error) {
		var a A
		a, err = io.unsafeRun()
		if err == nil {
			return f(a)
		} else {
			return recover(err)
		}
	})
}

// Recover handles a potential error from IO. It does not fail itself.
func Recover[A any](io IO[A], recover func(err error) IO[A]) IO[A] {
	return Fold(io, Lift[A], recover)
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
