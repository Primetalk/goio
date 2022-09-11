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

// TakeWhile returns the beginning of the stream such that all elements satisfy the predicate.
func TakeWhile[A any](stm Stream[A], predicate func(A) bool) Stream[A] {
	return Stream[A](io.Map(
		io.IO[StepResult[A]](stm),
		func(sra StepResult[A]) StepResult[A] {
			if sra.IsFinished || (sra.HasValue && !predicate(sra.Value)) {
				sra.IsFinished = true
			} else {
				sra.Continuation = TakeWhile(sra.Continuation, predicate)
			}
			return sra
		}))
}

// DropWhile removes the beginning of the stream so that the new stream starts with an element
// that falsifies the predicate.
func DropWhile[A any](stm Stream[A], predicate func(A) bool) Stream[A] {
	return Stream[A](io.Map(
		io.IO[StepResult[A]](stm),
		func(sra StepResult[A]) StepResult[A] {
			if !sra.IsFinished && (sra.HasValue && predicate(sra.Value)) {
				sra.HasValue = false
				sra.Continuation = DropWhile(sra.Continuation, predicate)
			}
			return sra
		}))
}
