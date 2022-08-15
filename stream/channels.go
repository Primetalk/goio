package stream

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/slice"
)

// ToChannel sends all stream elements to the given channel.
// When stream is completed, channel is closed.
// The IO blocks until the stream is exhausted.
// If the stream is failed, the channel is closed
func ToChannel[A any](stm Stream[A], ch chan<- A) io.IO[fun.Unit] {
	stmUnits := MapEval(stm,
		func(a A) io.IO[fun.Unit] {
			return io.FromPureEffect(func() {
				ch <- a
			})
		})
	return io.Finally(DrainAll(stmUnits), io.CloseChannel(ch))
}

// ToChannels sends each stream element to every given channel.
func ToChannels[A any](stm Stream[A], channels ...chan<- A) io.IO[fun.Unit] {
	stmUnits := MapEval(stm,
		func(a A) io.IO[fun.Unit] {
			return io.FromPureEffect(func() {
				for _, ch := range channels {
					ch <- a
				}
			})
		})
	return io.Finally(DrainAll(stmUnits),
		io.Map(
			io.Parallel(
				slice.Map(channels, io.CloseChannel[A]),
			),
			fun.Const[[]fun.Unit](fun.Unit1),
		),
	)
}

// FromChannel constructs a stream that reads from the given channel
// until the channel is open.
// When channel is closed, the stream is also closed.
func FromChannel[A any](ch <-chan A) Stream[A] {
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

// PairOfChannelsToPipe - takes two channels that are being used to
// talk to some external process and convert them into a single pipe.
// It first starts a separate go routine that will continuously run
// the input stream and send all it's contents to the `input` channel.
// The current thread is left with reading from the output channel.
func PairOfChannelsToPipe[A any, B any](input chan A, output chan B) Pipe[A, B] {
	return func(stmA Stream[A]) Stream[B] {
		return FlatMap(
			Eval(io.FireAndForget(ToChannel(stmA, input))),
			func(fun.Unit) Stream[B] {
				return FromChannel(output)
			})
	}
}

// PipeToPairOfChannels converts a streaming pipe to a pair of channels that could be used
// to interact with external systems.
func PipeToPairOfChannels[A any, B any](pipe Pipe[A, B]) io.IO[fun.Pair[chan<- A, <-chan B]] {
	return io.Delay(func() io.IO[fun.Pair[chan<- A, <-chan B]] {

		input := make(chan A)
		output := make(chan B)
		inputStream := FromChannel(input)
		outputStream := pipe(inputStream)

		return io.AndThen(
			io.FireAndForget(ToChannel(outputStream, output)),
			io.Lift(fun.Pair[chan<- A, <-chan B]{V1: input, V2: output}),
		)
	})
}
