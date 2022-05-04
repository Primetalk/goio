package stream

import "github.com/primetalk/goio/io"

type Stream[A any] interface {
	// step performs a single step in the state machine.
	// It might not 
	Step() (io.IO[StepResult[A]])
	IsFinished() io.IO[bool]
}





type StepResult[A any] struct {
	Value A
	Continuation Stream[A]
	HasValue bool
}




func Eval[A any](io io.IO[A]) Stream[A] {
	return evalImpl[A]{
		io: io,
	}
}

type evalImpl[A any] struct {
	io io.IO[A]
}

func (e evalImpl[A])Step() (io.IO[StepResult[A]]) {
	return io.MapPure(e.io, func(a A) StepResult[A]{
		return StepResult[A]{
			Value: a,
			Continuation: Empty[A](),
			HasValue: true,
		}
	})
}
func (e evalImpl[A])IsFinished() io.IO[bool] { return io.Lift(false) }



func MapEval[A any, B any](stm Stream[A], f func(a A)io.IO[B]) Stream[B] {
	return mapEvalImpl[A, B]{
		stm: stm,
		f: f,
	}
}

type mapEvalImpl[A any, B any] struct {
	stm Stream[A]
	f func(a A)io.IO[B]
}

func (e mapEvalImpl[A, B])Step() (io.IO[StepResult[B]]) {
	return io.FlatMap(
		e.stm.Step(), 
		func(sra StepResult[A]) io.IO[StepResult[B]] {
			if sra.HasValue {
				iob := e.f(sra.Value)
				return io.MapPure(iob, func (b B) StepResult[B]{
					return StepResult[B] {
						Value: b,
						HasValue: true,
						Continuation: MapEval(sra.Continuation, e.f),
					}
				})
			} else {
				return io.Lift(StepResult[B]{
					Continuation: MapEval(sra.Continuation, e.f),
					HasValue: false,
				})
			}
		})
}
func (e mapEvalImpl[A, B]) IsFinished() io.IO[bool] { return e.stm.IsFinished() }

func MapPure[A any, B any](stm Stream[A], f func(a A)B) Stream[B] {
	return MapEval(stm, func (a A) io.IO[B]{return io.Lift(f(a))})
}



func AndThenLazy[A any](stm1 Stream[A], stm2 func() Stream[A]) Stream[A] {
	return andThenImpl[A]{
		stm1: stm1,
		stm2: stm2,
	}
}


func AndThen[A any](stm1 Stream[A], stm2 Stream[A]) Stream[A] {
	return andThenImpl[A]{
		stm1: stm1,
		stm2: func() Stream[A] {return stm2},
	}
}

type andThenImpl[A any] struct {
	stm1 Stream[A]
	stm2 func() Stream[A]
}


func (a andThenImpl[A])Step() (io.IO[StepResult[A]]) {
	return  io.FlatMap(a.stm1.IsFinished(), func (stm1Finished bool) io.IO[StepResult[A]]{
		if stm1Finished {
			return a.stm2().Step()
		} else {
			return io.MapPure(a.stm1.Step(), func (sra StepResult[A]) StepResult[A] {
				return StepResult[A]{
					Value: sra.Value,
					Continuation: AndThenLazy(sra.Continuation, a.stm2),
					HasValue: sra.HasValue,
				}
			})
		}
	}) 
	
}

func (a andThenImpl[A])IsFinished() io.IO[bool] { 
	return io.FlatMap(a.stm1.IsFinished(), func (stm1Finished bool) io.IO[bool]{
		if stm1Finished {
			return a.stm2().IsFinished()
		} else {
			return io.Lift(false)
		}
	}) 
}








func FlatMap[A any, B any](stm Stream[A], f func (a A) Stream[B]) Stream[B] {
	return flatMapEvalImpl[A, B]{
		stm: stm,
		f: f,
	}
}

type flatMapEvalImpl[A any, B any] struct {
	stm Stream[A]
	f func (a A) Stream[B]
}

func (e flatMapEvalImpl[A, B])Step() (io.IO[StepResult[B]]) {
	return io.FlatMap(
		e.stm.Step(), 
		func(sra StepResult[A]) io.IO[StepResult[B]] {
			if sra.HasValue {
				stmb1 := e.f(sra.Value)
				stmb := AndThenLazy(stmb1, func() Stream[B]{return FlatMap(sra.Continuation, e.f)})
				return stmb.Step()
			} else {
				return io.Lift(
					StepResult[B]{
						HasValue: sra.HasValue,
						Continuation: FlatMap(sra.Continuation, e.f),
					})
				
			}
		})
}
func (e flatMapEvalImpl[A, B]) IsFinished() io.IO[bool] { return e.stm.IsFinished() }



// StateFlatMap maintains state along the way
func StateFlatMap[A any, B any, S any](stm Stream[A], zero S, f func (a A, s S) (S, Stream[B])) Stream[B] {
	return stateFlatMapImpl[A, B, S]{
		stm: stm,
		zero: zero,
		f: f,
	}
}


type stateFlatMapImpl[A any, B any, S any] struct {
	stm Stream[A]
	zero S
	f func (a A, s S) (S, Stream[B])
}

func (e stateFlatMapImpl[A, B, S])Step() (io.IO[StepResult[B]]) {
	return io.FlatMap(
		e.stm.Step(), 
		func(sra StepResult[A]) io.IO[StepResult[B]] {
			if sra.HasValue {
				st, stmb1 := e.f(sra.Value, e.zero)
				stmb := AndThenLazy(stmb1, func() Stream[B]{return StateFlatMap(sra.Continuation, st, e.f)})
				return stmb.Step()
			} else {
				return io.Lift(
					StepResult[B]{
						HasValue: sra.HasValue,
						Continuation: StateFlatMap(sra.Continuation, e.zero, e.f),
					})
				
			}
		})
}
func (e stateFlatMapImpl[A, B, S]) IsFinished() io.IO[bool] { return e.stm.IsFinished() }


