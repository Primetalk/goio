package stream

import "github.com/primetalk/goio/io"

type Pipe[A any, B any] func (Stream[A]) Stream[B]

type Sink[A any] Pipe[A, io.Unit]

func Through[A any, B any](stm Stream[A], pipe Pipe[A, B]) Stream[B] {
	return pipe(stm)
}

func NewSink[A any](f func(a A)) Sink[A] {
	return func (stm Stream[A]) Stream[io.Unit] {
		return Map(stm, func(a A) io.Unit {
			f(a)
			return io.Unit1
		})
	}
}

func ToSink[A any](stm Stream[A], sink Sink[A]) Stream[io.Unit] {
	return sink(stm)
}
