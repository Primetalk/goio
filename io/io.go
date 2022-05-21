package io

import (
	"github.com/pkg/errors"
	"github.com/primetalk/goio/fun"
)

type IO[A any] interface {
	unsafeRun() (A, error)
}

type GoResult[A any] struct {
	Value A
	Error error
}

func (e GoResult[A])unsafeRun() (res A, err error) {
	defer RecoverToErrorVar("GoResult.unsafeRun", &err)
	return e.Value, e.Error
}


func LiftPair[A any](a A, err error) IO[A] {
	return GoResult[A]{
		Value: a,
		Error: err,
	}
}

func UnsafeRunSync[A any](io IO[A]) (res A, err error) {
	defer RecoverToErrorVar("UnsafeRunSync", &err)
	return io.unsafeRun()
}

func Delay[A any](f func()IO[A]) IO[A] {
	return delayImpl[A]{
		f: f,
	}
}
type delayImpl[A any] struct {
	f func()IO[A]
}

func (i delayImpl[A])unsafeRun()(a A, err error) {
	defer RecoverToErrorVar("Delay.unsafeRun", &err)
	return i.f().unsafeRun()
}

func Eval[A any](f func () (A, error)) IO[A] {
	return evalImpl[A]{
		f: f,
	}
}


type evalImpl[A any] struct {
	f func () (A, error)
}

func (e evalImpl[A])unsafeRun() (res A, err error) {
	defer RecoverToErrorVar("Eval.unsafeRun", &err)
	return e.f()
}

func FromUnit(f func() error) IO[fun.Unit] {
	return Eval(func () (fun.Unit, error) {
		return fun.Unit1, f()
	})
}

func Pure[A any](f func() A) IO[A] {
	return Eval(func () (A, error) {
		return f(), nil
	})
}


func MapErr[A any, B any](ioA IO[A], f func(a A)(B, error)) IO[B] {
	return mapImpl[A, B]{
		io: ioA,
		f: f,
	}
}

func Map[A any, B any](ioA IO[A], f func(a A)B) IO[B] {
	return mapImpl[A, B]{
		io: ioA,
		f: func (a A)(B, error){ return f(a), nil},
	}
}

type mapImpl[A any, B any] struct {
	io IO[A]
	f func(a A)(B, error)
}

func (e mapImpl[A, B])unsafeRun() (res B, err error) {
	defer RecoverToErrorVar("Map.unsafeRun", &err)
	var a A
	a, err = e.io.unsafeRun()
	if err == nil {
		res, err = e.f(a)
	}
	return
}

func FlatMap[A any, B any](ioA IO[A], f func(a A) IO[B]) IO[B] {
	return flatMapImpl[A, B]{
		io: ioA,
		f: f,
	}
}

type flatMapImpl[A any, B any] struct {
	io IO[A]
	f func(a A) IO[B]
}

func (e flatMapImpl[A, B])unsafeRun() (res B, err error) {
	defer RecoverToErrorVar("FlatMap.unsafeRun", &err)
	var a A
	a, err = e.io.unsafeRun()
	if err == nil {
		res, err = e.f(a).unsafeRun()
	}
	return
}

func FlatMapErr[A any, B any](ioA IO[A], f func(a A) (B, error)) IO[B] {
	return flatMapImpl[A, B]{
		io: ioA,
		f: func(a A) IO[B] { return LiftPair(f(a))},
	}
}

// AndThen runs the first IO, ignores it's result and then runs the second one.
func AndThen[A any, B any](ioa IO[A], iob IO[B]) IO[B] {
	return FlatMap(ioa, func(A)IO[B]{
		return iob
	})
}
func Lift[A any](a A) IO[A] {
	return LiftPair(a, nil)
}

func Fail[A any](err error) IO[A] {
	var a A
	return LiftPair(a, err)
}

func Fold[A any, B any](io IO[A], f func(a A)IO[B], recover func (error)IO[B]) IO[B]{
	return Delay(func () IO[B] {
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
			return Lift(GoResult[A]{Value:a})
		},
		func(err error) IO[GoResult[A]] {
			return Lift(GoResult[A]{Error:err})
		},
	)
}

// UnfoldGoResult represents GoResult back to ordinary IO.
func UnfoldGoResult[A any](iogr IO[GoResult[A]]) IO[A] {
	return MapErr(iogr, func(gr GoResult[A]) (A, error) { return gr.Value, gr.Error})
}

func FoldErr[A any, B any](io IO[A], f func(a A) (B, error), recover func (error)(B, error)) IO[B]{
	return Eval(func () (b B, err error) {
		var a A
		a, err = io.unsafeRun()
		if err == nil {
			return f(a)			
		} else {
			return recover(err)
		}
	})
}

func Recover[A any](io IO[A], recover func(err error)IO[A])IO[A] {
	return Fold(io, Lift[A], recover)
}

func Sequence[A any](ioas []IO[A]) (res IO[[]A]) {
	res = Lift([]A{})
	for _, iou := range ioas {
		res = FlatMap(res, func (as []A) IO[[]A] { 
			return Map(iou, func (a A) []A { return append(as, a)})
		})
	}
	return
}

func SequenceUnit(ious []IOUnit) (res IOUnit) {
	res = IOUnit1
	for _, iou := range ious {
		res = FlatMap(res, func (fun.Unit) IOUnit { return iou })
	}
	return
}

// Unptr retrieves the value at pointer. Fails if nil
func Unptr[A any](ptra *A) IO[A]{
	if ptra == nil {
		return Fail[A](errors.New("nil pointer"))
	} else {
		return Lift(*ptra)
	}
}

// Wrapf wraps an error with additional context information
func Wrapf[A any](io IO[A], format string, args...interface{}) IO[A] {
	return Recover(io, func(err error) IO[A] {
		return Fail[A](errors.Wrapf(err, format, args...))
	})
}

var IOUnit1 = Lift(fun.Unit1)

type IOUnit = IO[fun.Unit]

// ForEach calls the provided callback after IO is completed.
func ForEach[A any](io IO[A], cb func(a A))IO[fun.Unit] {
	return Map(io, func (a A) fun.Unit {
		cb(a)
		return fun.Unit1
	})
}
