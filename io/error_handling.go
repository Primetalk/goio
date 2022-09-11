package io

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/option"
)

// FoldToGoResult converts either value or error to go result
// typically it should never fail.
func FoldToGoResult[A any](io IO[A]) IO[GoResult[A]] {
	return Fold(
		io,
		func(a A) IO[GoResult[A]] {
			return Lift(GoResult[A]{Value: a})
		},
		func(err error) IO[GoResult[A]] {
			return Lift(GoResult[A]{Error: err})
		},
	)
}

// UnfoldGoResult represents GoResult back to ordinary IO.
func UnfoldGoResult[A any](iogr IO[GoResult[A]]) IO[A] {
	return MapErr(iogr, func(gr GoResult[A]) (A, error) { return gr.Value, gr.Error })
}

// Recover handles a potential error from IO. It does not fail itself.
func Recover[A any](io IO[A], recover func(err error) IO[A]) IO[A] {
	return Fold(io, Lift[A], recover)
}

// OnError executes a side effect when there is an error.
func OnError[A any](io IO[A], onError func(err error) IO[fun.Unit]) IO[A] {
	return Fold(io, Lift[A], func(err error) IO[A] {
		return AndThen(onError(err), Fail[A](err))
	})
}

// Retry performs the same operation a few times based on the retry strategy.
func Retry[A any, S any](ioa IO[A], strategy func(s S, err error) IO[option.Option[S]], zero S) IO[A] {
	return Recover(ioa, func(err error) IO[A] {
		return FlatMap(strategy(zero, err), func(os option.Option[S]) IO[A] {
			return option.Fold(os,
				func(s S) IO[A] {
					return Retry(ioa, strategy, s)
				},
				func() IO[A] {
					return Fail[A](err)
				},
			)
		})
	})
}

// RetryS performs the same operation a few times based on the retry strategy.
// Also returns the last state of the error-handling strategy.
func RetryS[A any, S any](ioa IO[A], strategy func(s S, err error) IO[option.Option[S]], zero S) IO[fun.Pair[A, S]] {
	return Recover(
		Map(ioa, func(a A) fun.Pair[A, S] { return fun.NewPair(a, zero) }),
		func(err error) IO[fun.Pair[A, S]] {
			return FlatMap(strategy(zero, err), func(os option.Option[S]) IO[fun.Pair[A, S]] {
				return option.Fold(os,
					func(s S) IO[fun.Pair[A, S]] {
						return RetryS(ioa, strategy, s)
					},
					func() IO[fun.Pair[A, S]] {
						return Fail[fun.Pair[A, S]](err)
					},
				)
			})
		})
}

// RetryStrategyMaxCount is a strategy that retries n times immediately.
func RetryStrategyMaxCount(substring string) func(s int, err error) IO[option.Option[int]] {
	return func(s int, err error) IO[option.Option[int]] {
		return Pure(func() option.Option[int] {
			if s <= 0 {
				return option.None[int]()
			} else {
				return option.Some(s - 1)
			}
		})
	}
}
