package fstream

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/option"
	"github.com/primetalk/goio/stream"
)

// Collect collects all element from the stream and for each element invokes
// the provided function
func Collect[A any](stm Stream[A], collector func(A) error) io.IOUnit {
	return io.AndThen(
		io.Finally(stream.Collect(stm.init.sa, collector), stm.init.finalizer),
		option.Match(stm.cont,
			func(iosa2 IOStream[A]) io.IOUnit {
				return io.FlatMap(io.IO[Stream[A]](iosa2), func(sa2 Stream[A]) io.IOUnit {
					return Collect(sa2, collector)
				})
			},
			fun.Delay(io.IOUnit1),
		),
	)
}

func returnNilError[A any](collector func(A)) func(A) error {
	return func(a A) error {
		collector(a)
		return nil
	}
}

// ForEach invokes a simple function for each element of the stream.
func ForEach[A any](stm Stream[A], collector func(A)) io.IO[fun.Unit] {
	return Collect(stm, returnNilError(collector))
}

// DrainAll executes the stream and throws away all values.
func DrainAll[A any](stm Stream[A]) io.IO[fun.Unit] {
	return Collect(stm, fun.Const[A, error](nil))
}

// AppendToSlice executes the stream and appends it's results to the slice.
func AppendToSlice[A any](stm Stream[A], start []A) io.IO[[]A] {
	appendToStart := func(a A) error {
		start = append(start, a)
		return nil
	}
	return io.AndThen(
		Collect(stm, appendToStart),
		io.Delay(func() io.IO[[]A] {
			return io.Lift(start)
		}),
	)
}

// ToSlice executes the stream and collects all results to a slice.
func ToSlice[A any](stm Stream[A]) io.IO[[]A] {
	return AppendToSlice(stm, []A{})
}

// Head takes the first element and returns it.
// It'll fail if the stream is empty.
func Head[A any](stm Stream[A]) io.IO[A] {
	res := StreamMatch[A, A](
		stm,
		/*onFinish*/ func() io.IO[A] {
			return io.Fail[A](stream.ErrHeadOfEmptyStream)
		},
		/*onValue */ func(a A, tail Stream[A]) io.IO[A] {
			return io.AndThen(tail.init.finalizer, io.Lift(a))
		},
		/*onEmpty */ func(tail Stream[A]) io.IO[A] {
			return Head(tail)
		},
		/*onError */ func(err error) io.IO[A] {
			return io.Fail[A](err)
		},
	)
	return res
}

// IOHead takes the first element of io-stream and returns it.
// It'll fail if the stream is empty.
func IOHead[A any](stm IOStream[A]) io.IO[A] {
	return io.FlatMap[Stream[A]](io.IO[Stream[A]](stm), Head[A])
}

// Last keeps track of the current element of the stream
// and returns it when the stream completes.
func Last[A any](stm Stream[A]) io.IO[A] {
	ea := EmptyIO[A]()
	return IOHead(StateFlatMapWithFinish[A, A, option.Option[A]](
		stm,
		/*zero*/ option.None[A](),
		/*prefix*/ EmptyIO[A](),
		/*f*/ func(a A, s option.Option[A]) (option.Option[A], IOStream[A]) {
			return option.Some(a), ea
		},
		/*onFinish*/ func(s option.Option[A]) IOStream[A] {
			return option.Match(s, 
				LiftIO[A],
				EmptyIO[A],
			)
		},	
	))
}
