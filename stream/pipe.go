package stream

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
)

// Pipe is a conversion of one stream to another.
// Technically it's a function that takes one stream and returns another.
type Pipe[A any, B any] func(Stream[A]) Stream[B]

// Sink is a pipe that does not return meaningful values.
type Sink[A any] Pipe[A, fun.Unit]

// Through passes the stream data through the pipe.
// Technically it applies the pipe function to this stream.
func Through[A any, B any](stm Stream[A], pipe Pipe[A, B]) Stream[B] {
	return pipe(stm)
}

// ThroughPipeEval runs the given stream through pipe that is returned by
// the provided pipeIO.
func ThroughPipeEval[A any, B any](stm Stream[A], pipeIO io.IO[Pipe[A, B]]) Stream[B] {
	return FlatMap(Eval(pipeIO), func(pipe Pipe[A, B]) Stream[B] {
		return pipe(stm)
	})
}

// NewSink constructs the sink from the provided function.
func NewSink[A any](f func(a A)) Sink[A] {
	return func(stm Stream[A]) Stream[fun.Unit] {
		return Map(stm, func(a A) fun.Unit {
			f(a)
			return fun.Unit1
		})
	}
}

// ToSink streams all data from the stream to the sink.
func ToSink[A any](stm Stream[A], sink Sink[A]) Stream[fun.Unit] {
	return sink(stm)
}

// ConcatPipes connects two pipes into one.
func ConcatPipes[A any, B any, C any](pipe1 Pipe[A, B], pipe2 Pipe[B, C]) Pipe[A, C] {
	return func(sa Stream[A]) Stream[C] {
		return pipe2(pipe1(sa))
	}
}

// PrependPipeToSink changes the input of a sink.
func PrependPipeToSink[A any, B any](pipe1 Pipe[A, B], sink Sink[B]) Sink[A] {
	return Sink[A](ConcatPipes(pipe1, Pipe[B, fun.Unit](sink)))
}
