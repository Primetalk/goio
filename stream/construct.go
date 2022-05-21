package stream

import "github.com/primetalk/goio/io"



func Empty[A any]()Stream[A]{
	return io.Pure(func () StepResult[A] {return NewStepResultFinished[A]()})
}


func Eval[A any](ioa io.IO[A]) Stream[A] {
	return io.Map(ioa, func(a A) StepResult[A]{
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
	if len(as) == 0 {
		return io.Lift(NewStepResultFinished[A]())
	} else {
		return io.Lift(NewStepResult(as[0], FromSlice(as[1:])))
	}
}


func Generate[A any, S any](zero S, f func(s S) (S, A)) Stream[A] {
	return io.Eval(func()(StepResult[A], error){
		s, a := f(zero)
		return NewStepResult(a, Generate(s, f)), nil
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

func FromStepResult[A any](iosr io.IO[StepResult[A]]) Stream[A] {
	return iosr
}
