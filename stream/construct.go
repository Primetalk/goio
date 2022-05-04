package stream

import "github.com/primetalk/goio/io"



func Empty[A any]()Stream[A]{
	return emptyStream[A]{}
}

type emptyStream[A any] struct {}

func (emptyStream[A])Step() (io.IO[StepResult[A]]) {
	res := StepResult[A]{
		Continuation: Empty[A](),
		HasValue: false,
	}
	return io.Lift(res)
}

func (emptyStream[A])IsFinished() io.IO[bool] { return io.Lift(true) }





func Lift[A any](a A) Stream[A] {
	return Eval(io.Lift(a))
}

func LiftMany[A any](as ...A) Stream[A] {
	return FromSlice(as)
}


func FromSlice[A any](as []A) Stream[A] {
	return fromSliceImpl[A]{
		slice: as,
	}
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
		return io.Lift(StepResult[A]{
			Continuation: a,
			HasValue: false,
		})
	} else {
		return io.Lift(StepResult[A]{
			Value: a.slice[0],
			Continuation: FromSlice(a.slice[1:]),
			HasValue: true,
		})
	}
}

func (a fromSliceImpl[A])IsFinished() io.IO[bool] { 
	return io.Lift(len(a.slice) == 0)
}
