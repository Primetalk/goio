package stream

import "github.com/primetalk/goio/io"


func DrainAll[A any](stm Stream[A]) io.IO[io.Unit] {
	return io.FlatMap(
		stm.IsFinished(), 
		func (finished bool) io.IO[io.Unit] {
			if finished {
				return io.Lift(io.Unit{})
			} else {
				return io.FlatMap(stm.Step(), func(sra StepResult[A]) io.IO[io.Unit] {
					return DrainAll(sra.Continuation)
				})
			}
		})
}



func AppendToSlice[A any](stm Stream[A], start []A) io.IO[[]A] {
	return io.FlatMap(
		stm.IsFinished(), 
		func (finished bool) io.IO[[]A] {
			if finished {
				return io.Lift(start)
			} else {
				return io.FlatMap(stm.Step(), func(sra StepResult[A]) io.IO[[]A] {
					return AppendToSlice(sra.Continuation, append(start, sra.Value))
				})
			}
		})
}

func ToSlice[A any](stm Stream[A]) io.IO[[]A]{
	return AppendToSlice(stm, []A{})
}
