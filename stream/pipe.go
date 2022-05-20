package stream

import "github.com/primetalk/goio/fun"

type Pipe[A any, B any] func(Stream[A]) Stream[B]

type Sink[A any] Pipe[A, fun.Unit]

func Through[A any, B any](stm Stream[A], pipe Pipe[A, B]) Stream[B] {
	return pipe(stm)
}

func NewSink[A any](f func(a A)) Sink[A] {
	return func(stm Stream[A]) Stream[fun.Unit] {
		return Map(stm, func(a A) fun.Unit {
			f(a)
			return fun.Unit1
		})
	}
}

func ToSink[A any](stm Stream[A], sink Sink[A]) Stream[fun.Unit] {
	return sink(stm)
}
