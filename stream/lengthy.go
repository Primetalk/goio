package stream

import "github.com/primetalk/goio/io"


func Repeat[A any](stm Stream[A]) Stream[A] {
	return AndThenLazy(stm, func() Stream[A] {return Repeat(stm)})
}

func Take[A any](stm Stream[A], n int) Stream[A]{
	if n <= 0 {
		return Empty[A]()
	} else {
		return takeImpl[A]{
			stm: stm,
			n: n,
		}
	}
}

type takeImpl[A any] struct {
	stm Stream[A]
	n int // invariant n > 0!
}

func (t takeImpl[A])Step() (io.IO[StepResult[A]]) {
	if t.n == 0 {// should never happen
		return Empty[A]().Step()
	} else {
		return io.MapPure(t.stm.Step(), func (sra StepResult[A]) StepResult[A] {
			sra.Continuation = Take(sra.Continuation, t.n - 1)
			return sra
		})
	}
}

func (t takeImpl[A])IsFinished() io.IO[bool] { return t.stm.IsFinished() }

func Drop[A any](stm Stream[A], n int) Stream[A]{
	if n <= 0 {
		return Empty[A]()
	} else {
		return dropImpl[A]{
			stm: stm,
			n: n,
		}
	}
}

type dropImpl[A any] struct {
	stm Stream[A]
	n int // invariant n > 0!
}

func (t dropImpl[A])Step() (io.IO[StepResult[A]]) {
	if t.n == 0 {// should never happen
		return Empty[A]().Step()
	} else {
		return io.MapPure(t.stm.Step(), func (sra StepResult[A]) StepResult[A] {
			sra.Continuation = Take(sra.Continuation, t.n - 1)
			sra.HasValue = false
			return sra
		})
	}
}

func (t dropImpl[A])IsFinished() io.IO[bool] { return t.stm.IsFinished() }
