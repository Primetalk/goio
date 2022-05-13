package stream

import "github.com/primetalk/goio/io"



func Empty[A any]()Stream[A]{
	return emptyStream[A]{}.Step
}

type emptyStream[A any] struct {}

func (es emptyStream[A])Step() (io.IO[StepResult[A]]) {
	return io.Lift(NewStepResultFinished[A]())
}


func Eval[A any](io io.IO[A]) Stream[A] {
	return evalImpl[A]{
		io: io,
	}.Step
}

type evalImpl[A any] struct {
	io io.IO[A]
}

func (e evalImpl[A])Step() (io.IO[StepResult[A]]) {
	return io.Map(e.io, func(a A) StepResult[A]{
		return NewStepResult(a, Empty[A]())
	})
}


func Lift[A any](a A) Stream[A] {
	return Eval(io.Lift(a))
}

func LiftMany[A any](as ...A) Stream[A] {
	return FromSlice(as)
}


func FromSlice[A any](as []A) Stream[A] {
	return fromSliceImpl[A]{
		slice: as,
	}.Step
	// if len(as) == 0 {
	// 	return Empty[A]()
	// } else if len(as) == 1 {
	// 	return Lift(as[0])
	// } else {
	// 	AndThen[A]()
	// }
	// return FromSlice(as)
}

type fromSliceImpl[A any] struct {
	slice []A
}


func (a fromSliceImpl[A])Step() (io.IO[StepResult[A]]) {
	if len(a.slice) == 0 {
		return io.Lift(NewStepResultFinished[A]())
	} else {
		return io.Lift(NewStepResult(a.slice[0], FromSlice(a.slice[1:])))
	}
}


func Generate[A any, S any](zero S, f func(s S) (S, A)) Stream[A] {
	return generateImpl[A, S]{
		zero: zero,
		f: f,
	}.Step
}

type generateImpl[A any, S any] struct {
	zero S
	f func(s S) (S, A)
}

func (g generateImpl[A, S])Step() (io.IO[StepResult[A]]) {
	return io.Eval(func()(StepResult[A], error){
		s, a := g.f(g.zero)
		return NewStepResult(a, Generate(s, g.f)), nil
	})
}

func Unfold[A any](zero A, f func(A) A) Stream[A] {
	return Generate(zero, func(s A)(A,A) {
		r := f(s)
		return r, r
	})
}

func FromSideEffectfulFunction[A any](f func ()(A,error)) Stream[A] {
	 return Repeat(Eval(io.Eval(f)))
}
