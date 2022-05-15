package stream

import "github.com/primetalk/goio/io"


func Repeat[A any](stm Stream[A]) Stream[A] {
	return AndThenLazy(stm, func() Stream[A] {return Repeat(stm)})
}

func Take[A any](stm Stream[A], n int) Stream[A]{
	if n <= 0 {
		return Empty[A]()
	} else {
		return io.Map[StepResult[A]](stm, func (sra StepResult[A]) StepResult[A] {
			nextCount := n
			if sra.HasValue {
				nextCount = n - 1
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
		return io.Map[StepResult[A]](stm, func (sra StepResult[A]) StepResult[A] {
			sra.Continuation = Drop(sra.Continuation, n - 1)
			sra.HasValue = false
			return sra
		})
	}
}
