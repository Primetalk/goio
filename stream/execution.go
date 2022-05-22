package stream

import (
	"fmt"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
)

// Collect collects all element from the stream and for each element invokes
// the provided function
func Collect[A any](stm Stream[A], collector func(A) error) io.IO[fun.Unit] {
	return io.FlatMap[StepResult[A]](
		stm,
		func(sra StepResult[A]) io.IO[fun.Unit] {
			if sra.IsFinished {
				return io.Lift(fun.Unit1)
			} else {
				if sra.HasValue {
					collector(sra.Value)
				}
				return Collect(sra.Continuation, collector)
			}
		})
}

// ForEach invokes a simple function for each element of the stream.
func ForEach[A any](stm Stream[A], collector func(A)) io.IO[fun.Unit] {
	return Collect(stm, func(a A) error {
		collector(a)
		return nil
	})
}

// DrainAll executes the stream and throws away all values.
func DrainAll[A any](stm Stream[A]) io.IO[fun.Unit] {
	return io.FlatMap[StepResult[A]](stm, func(sra StepResult[A]) io.IO[fun.Unit] {
		if sra.IsFinished {
			return io.Lift(fun.Unit1)
		} else {
			return DrainAll(sra.Continuation)
		}
	})
}

// AppendToSlice executes the stream and appends it's results to the slice.
func AppendToSlice[A any](stm Stream[A], start []A) io.IO[[]A] {
	return io.FlatMap[StepResult[A]](stm, func(sra StepResult[A]) io.IO[[]A] {
		if sra.IsFinished {
			return io.Lift(start)
		} else if sra.HasValue {
			return AppendToSlice(sra.Continuation, append(start, sra.Value))
		} else {
			return AppendToSlice(sra.Continuation, start)
		}
	})
}

// ToSlice executes the stream and collects all results to a slice.
func ToSlice[A any](stm Stream[A]) io.IO[[]A] {
	return AppendToSlice(stm, []A{})
}

// Head takes the first element and executes it.
// It'll fail if the stream is empty.
func Head[A any](stm Stream[A]) io.IO[A] {
	slice1 := ToSlice(Take(stm, 1))
	return io.MapErr(slice1, func(as []A) (a A, err error) {
		if len(as) > 0 {
			a = as[0]
		} else {
			err = fmt.Errorf("head of an empty stream")
		}
		return
	})
}

// ToChannel sends all stream elements to the given channel.
// When stream is completed, channel is closed.
func ToChannel[A any](stm Stream[A], ch chan A) io.IO[fun.Unit] {
	stmUnits := StateFlatMapWithFinish(stm, ch, 
		func(a A, ch chan A) (chan A, Stream[fun.Unit]){
			ch <- a
			return ch, EmptyUnit()
		}, 
		func(ch chan A) Stream[fun.Unit] {
			close(ch)
			return EmptyUnit()
		})
	return DrainAll(stmUnits)
}

// FromChannel constructs a stream that reads from the given channel
// until the channel is open.
// When channel is closed, the stream is also closed.
func FromChannel[A any](ch chan A) Stream[A] {
	return FromStepResult(
		io.Pure(func() StepResult[A] {
			a, ok := <-ch
			if ok {
				return NewStepResult(a, FromChannel(ch))
			} else {
				return NewStepResultFinished[A]()
			}
		}),
	)
}
