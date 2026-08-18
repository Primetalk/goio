package io

import (
	"errors"
	"time"

	"github.com/primetalk/goio/fun"
)

// Sleep makes the IO sleep the specified time.
func Sleep(d time.Duration) IO[fun.Unit] {
	return FromPureEffect(func() {
		time.Sleep(d)
	})
}

// SleepA sleeps and then returns the constant value.
func SleepA[A any](d time.Duration, value A) IO[A] {
	return Map(Sleep(d), fun.ConstUnit(value))
}

// ErrorTimeout is an error that will be returned in case of timeout.
var ErrorTimeout = errors.New("timeout")

// WithTimeout returns the IO result if it completes within the provided
// duration, or fails with ErrorTimeout when the timeout wins.
//
// WithTimeout only stops waiting. It does not cancel the input IO, which
// continues independently and may still perform side effects after timeout.
func WithTimeout[A any](d time.Duration) func(ioa IO[A]) IO[A] {
	return func(ioa IO[A]) IO[A] {
		first := ConcurrentlyFirst([]IO[GoResult[A]]{
			FoldToGoResult(ioa),
			SleepA(d, GoResult[A]{Error: ErrorTimeout}),
		})
		return UnfoldGoResult(first)
	}
}

// Never is a simple IO that never returns.
func Never[A any]() IO[A] {
	return Async(func(c Callback[A]) {})
}

// Notify starts a separate thread that will call the given callback after
// the specified time.
func Notify[A any](d time.Duration, value A, cb Callback[A]) IO[fun.Unit] {
	return FireAndForget(
		ForEach(
			SleepA(d, value),
			func(a A) {
				cb(a, nil)
			}),
	)
}

// NotifyToChannel sends message to channel after specified duration.
func NotifyToChannel[A any](d time.Duration, value A, ch chan A) IO[fun.Unit] {
	return Notify(d, value, func(str A, err error) {
		ch <- value
	})
}

// AfterTimeout sleeps the given time and then starts the other IO.
func AfterTimeout[A any](duration time.Duration, ioa IO[A]) IO[A] {
	return AndThen(Sleep(duration), ioa)
}
