package stream

import "github.com/primetalk/goio/io"

// Repeat appends the same stream infinitely.
func Repeat[A any](stm Stream[A]) Stream[A] {
	return AndThenLazy(stm, func() Stream[A] { return Repeat(stm) })
}

// Take cuts the stream after n elements.
func Take[A any](stm Stream[A], n int) Stream[A] {
	if n <= 0 {
		return Empty[A]()
	} else {
		return Stream[A](io.Map(
			io.IO[StepResult[A]](stm),
			func(sra StepResult[A]) StepResult[A] {
				nextCount := n
				if sra.HasValue {
					nextCount = n - 1
				}
				sra.Continuation = Take(sra.Continuation, nextCount)
				return sra
			}))
	}
}

// Drop skips n elements in the stream.
func Drop[A any](stm Stream[A], n int) Stream[A] {
	if n <= 0 {
		return stm
	} else {
		return Stream[A](io.Map(
			io.IO[StepResult[A]](stm),
			func(sra StepResult[A]) StepResult[A] {
				sra.Continuation = Drop(sra.Continuation, n-1)
				sra.HasValue = false
				return sra
			}))
	}
}
