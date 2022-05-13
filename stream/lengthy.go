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
		}.Step
	}
}

type takeImpl[A any] struct {
	stm Stream[A]
	n int // invariant n > 0!
}

func (t takeImpl[A])Step() (io.IO[StepResult[A]]) {
	if t.n == 0 {// should never happen
		return Empty[A]()()
	} else {
		return io.Map(t.stm(), func (sra StepResult[A]) StepResult[A] {
			nextCount := t.n
			if sra.HasValue {
				nextCount = t.n - 1
			}
			sra.Continuation = Take(sra.Continuation, nextCount)
			return sra
		})
	}
}

func Drop[A any](stm Stream[A], n int) Stream[A]{
	if n <= 0 {
		return stm
	} else {
		return dropImpl[A]{
			stm: stm,
			n: n,
		}.Step
	}
}

type dropImpl[A any] struct {
	stm Stream[A]
	n int // invariant n > 0!
}

func (i dropImpl[A])Step() (io.IO[StepResult[A]]) {
	if i.n == 0 {// should never happen
		return Empty[A]()()
	} else {
		return io.Map(i.stm(), func (sra StepResult[A]) StepResult[A] {
			sra.Continuation = Drop(sra.Continuation, i.n - 1)
			sra.HasValue = false
			return sra
		})
	}
}
