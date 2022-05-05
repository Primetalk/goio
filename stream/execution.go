package stream

import (
	"fmt"

	"github.com/primetalk/goio/io"
)


func DrainAll[A any](stm Stream[A]) io.IO[io.Unit] {
	return io.FlatMap(stm.Step(), func(sra StepResult[A]) io.IO[io.Unit] {
		if sra.IsFinished {
			return io.Lift(io.Unit1)
		} else {
			return DrainAll(sra.Continuation)
		}
	})
}



func AppendToSlice[A any](stm Stream[A], start []A) io.IO[[]A] {
	return io.FlatMap(stm.Step(), func(sra StepResult[A]) io.IO[[]A] {
		if sra.IsFinished {
			return io.Lift(start)
		} else if sra.HasValue {
			return AppendToSlice(sra.Continuation, append(start, sra.Value))
		} else {
			return AppendToSlice(sra.Continuation, start)
		}
	})
}

func ToSlice[A any](stm Stream[A]) io.IO[[]A]{
	return AppendToSlice(stm, []A{})
}

func Head[A any](stm Stream[A]) io.IO[A] {
	slice1 := ToSlice(Take(stm, 1))
	return io.MapErr(slice1, func (as []A) (a A, err error) {
		if len(as) > 0 {
			a = as[0]
		} else {
			err = fmt.Errorf("head of empty stream")
		}
		return
	})
}
