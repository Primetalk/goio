package stream

import "github.com/primetalk/goio/io"

// Stream is modelled as a function that performs a single step in the state machine.
type Stream[A any] func() (io.IO[StepResult[A]])


type StepResult[A any] struct {
	Value A
	Continuation Stream[A]
	HasValue bool
	IsFinished bool
}

func NewStepResult[A any](value A, continuation Stream[A]) StepResult[A] {
	return StepResult[A]{
		Value: value,
		HasValue: true,
		Continuation: continuation,
		IsFinished: false,
	}
}

func NewStepResultEmpty[A any](continuation Stream[A]) StepResult[A] {
	return StepResult[A]{
		HasValue: false,
		Continuation: continuation,
		IsFinished: false,
	}
}

func NewStepResultFinished[A any]() StepResult[A] {
	return StepResult[A]{
		IsFinished: true,
		Continuation: Empty[A](),
	}
}



func MapEval[A any, B any](stm Stream[A], f func(a A)io.IO[B]) Stream[B] {
	return mapEvalImpl[A, B]{
		stm: stm,
		f: f,
	}.Step
}

type mapEvalImpl[A any, B any] struct {
	stm Stream[A]
	f func(a A)io.IO[B]
}

func (e mapEvalImpl[A, B])Step() (io.IO[StepResult[B]]) {
	return io.FlatMap(
		e.stm(), 
		func(sra StepResult[A]) io.IO[StepResult[B]] {
			if sra.IsFinished {
				return io.Lift(NewStepResultFinished[B]())
			} else if sra.HasValue {
				iob := e.f(sra.Value)
				return io.Map(iob, func (b B) StepResult[B] {
					return NewStepResult(b, MapEval(sra.Continuation, e.f))
				})
			} else {
				return io.Lift(
					NewStepResultEmpty(MapEval(sra.Continuation, e.f)),
				)
			}
		})
}

func Map[A any, B any](stm Stream[A], f func(a A)B) Stream[B] {
	return MapEval(stm, func (a A) io.IO[B]{return io.Lift(f(a))})
}



func AndThenLazy[A any](stm1 Stream[A], stm2 func() Stream[A]) Stream[A] {
	return andThenImpl[A]{
		stm1: stm1,
		stm2: stm2,
	}.Step
}


func AndThen[A any](stm1 Stream[A], stm2 Stream[A]) Stream[A] {
	return andThenImpl[A]{
		stm1: stm1,
		stm2: func() Stream[A] {return stm2},
	}.Step
}

type andThenImpl[A any] struct {
	stm1 Stream[A]
	stm2 func() Stream[A]
}


func (a andThenImpl[A])Step() (io.IO[StepResult[A]]) {
	return  io.FlatMap(a.stm1(), func (sra StepResult[A]) io.IO[StepResult[A]]{
		if sra.IsFinished {
			return a.stm2()()
		} else {
			return io.Lift(StepResult[A]{
				Value: sra.Value,
				Continuation: AndThenLazy(sra.Continuation, a.stm2),
				HasValue: sra.HasValue,
			})
		}
	}) 
	
}






func FlatMap[A any, B any](stm Stream[A], f func (a A) Stream[B]) Stream[B] {
	return flatMapEvalImpl[A, B]{
		stm: stm,
		f: f,
	}.Step
}

type flatMapEvalImpl[A any, B any] struct {
	stm Stream[A]
	f func (a A) Stream[B]
}

func (e flatMapEvalImpl[A, B])Step() (io.IO[StepResult[B]]) {
	return io.FlatMap(
		e.stm(), 
		func(sra StepResult[A]) io.IO[StepResult[B]] {
			if sra.IsFinished {
				return io.Lift(NewStepResultFinished[B]())
			} else if sra.HasValue {
				stmb1 := e.f(sra.Value)
				stmb := AndThenLazy(stmb1, func() Stream[B]{return FlatMap(sra.Continuation, e.f)})
				return stmb()
			} else {
				return io.Lift(NewStepResultEmpty(FlatMap(sra.Continuation, e.f)))
			}
		})
}


// StateFlatMap maintains state along the way
func StateFlatMap[A any, B any, S any](stm Stream[A], zero S, f func (a A, s S) (S, Stream[B])) Stream[B] {
	return stateFlatMapImpl[A, B, S]{
		stm: stm,
		zero: zero,
		f: f,
		onFinish: func (S) Stream[B] { return Empty[B]()},
	}.Step
}


type stateFlatMapImpl[A any, B any, S any] struct {
	stm Stream[A]
	zero S
	f func (a A, s S) (S, Stream[B])
	onFinish func (s S) (Stream[B])
}

func (e stateFlatMapImpl[A, B, S])Step() (io.IO[StepResult[B]]) {
	return io.FlatMap(
		e.stm(), 
		func(sra StepResult[A]) (iores io.IO[StepResult[B]]) {
			if sra.IsFinished {
				iores = io.Lift(NewStepResultEmpty(e.onFinish(e.zero)))
			} else if sra.HasValue {
				st, stmb1 := e.f(sra.Value, e.zero)
				stmb := AndThenLazy(stmb1, func() Stream[B]{return StateFlatMap(sra.Continuation, st, e.f)})
				iores = stmb()
			} else {
				iores = io.Lift(NewStepResultEmpty(StateFlatMap(sra.Continuation, e.zero, e.f)))
			}
			return
		})
}


func Filter[A any](stm Stream[A], f func(A)bool) Stream[A] {
	return filterImpl[A]{
		stm: stm,
		f: f,
	}.Step
}


type filterImpl[A any] struct {
	stm Stream[A]
	f func(A)bool
}

func (e filterImpl[A])Step() (io.IO[StepResult[A]]) {
	return io.Map(e.stm(),
		func (sra StepResult[A]) StepResult[A] {
			if sra.IsFinished {
				return NewStepResultFinished[A]()
			} else {
				cont := Filter(sra.Continuation, e.f)
				if sra.HasValue && e.f(sra.Value) {
					return NewStepResult(sra.Value, cont)
				} else {
					return NewStepResultEmpty(cont)
				}
			}
		})
}

