package io

import (
	"errors"
	"sync"
	"time"

	"github.com/primetalk/goio/fun"
)

// ErrorFiberClosed indicates that observation of a fiber was closed before its work completed.
// Closing observation does not cancel or stop the underlying work.
var ErrorFiberClosed = errors.New("fiber observation is closed")

// Fiber[A] is a type-safe handle for observing work running in a Go routine.
// Join returns the first terminal observation published by work completion or Close.
type Fiber[A any] interface {
	// Join waits for results of the fiber.
	// When work completes before observation is closed, Join returns its result.
	// When Close wins first, current and future joins fail with ErrorFiberClosed.
	Join() IO[A]
	// Close shuts down observation of an incomplete fiber.
	//
	// Close publishes ErrorFiberClosed to all current joiners and makes future
	// joins fail with the same error. Close is idempotent. If work completed
	// first, Close preserves the completed result. If Close completed first,
	// later work completion is ignored by this observation handle.
	//
	// Close is not cancellation: it sends no signal to the underlying work,
	// which continues independently and may still perform side effects.
	Close() IO[fun.Unit]
	// Cancel sends cancellation signal to the Fiber.
	// If the fiber respects the signal, it'll stop.
	// Yet to be implemented.
	// Cancel() IO[Unit]
}

// if result is already available, there is no need to use callbacks channel.
// The result will be immediately delivered.
type fiberImpl[A any] struct {
	mu        *sync.Mutex
	result    *GoResult[A]
	callbacks []Callback[A]
}

var _ Fiber[any] = (*fiberImpl[any])(nil)

func (f *fiberImpl[A]) registerJoiner(cb Callback[A]) {
	f.mu.Lock()
	if f.result == nil {
		f.callbacks = append(f.callbacks, cb)
		f.mu.Unlock()
		return
	}
	result := *f.result
	f.mu.Unlock()

	cb(result.Value, result.Error)
}

func (f *fiberImpl[A]) publishResult(result GoResult[A]) bool {
	f.mu.Lock()
	if f.result != nil {
		f.mu.Unlock()
		return false
	}
	f.result = &result
	callbacks := f.callbacks
	f.callbacks = nil
	f.mu.Unlock()

	for _, cb := range callbacks {
		cb(result.Value, result.Error)
	}
	return true
}

func (f *fiberImpl[A]) Join() IO[A] {
	return Async(f.registerJoiner)
}

func (f *fiberImpl[A]) Close() IO[fun.Unit] {
	return FromPureEffect(func() {
		f.publishResult(GoResult[A]{Error: ErrorFiberClosed})
	})
}

// StartInExecutionContext executes the given task in the provided ExecutionContext
// It'll establish a channel with callbacks, so that
// any number of listeners could join the returned fiber. (Simultaneously not more than MaxCallbackCount though.)
// When completed it'll start sending the results to the callbacks.
// The same value will be delivered to all listeners.
func StartInExecutionContext[A any](ec ExecutionContext) func(io IO[A]) IO[Fiber[A]] {
	return func(io IO[A]) IO[Fiber[A]] {
		return Delay(func() IO[Fiber[A]] {
			fiber := &fiberImpl[A]{
				mu:        &sync.Mutex{},
				callbacks: []Callback[A]{},
			}
			goRoutine := func() {
				defer fun.RecoverToLog("StartInExecutionContext.goRoutine")
				a, err1 := UnsafeRunSync(io)
				fiber.publishResult(GoResult[A]{Value: a, Error: err1})
			}
			return Map(ec.Start(goRoutine), fun.ConstUnit[Fiber[A]](fiber))
		})
	}
}

// Start will start the IO in a separate go-routine (actually in the global unbounded execution context).
// It'll establish a channel with callbacks, so that
// any number of listeners could join the returned fiber.
// When completed it'll start sending the results to the callbacks.
// The same value will be delivered to all listeners.
func Start[A any](io IO[A]) IO[Fiber[A]] {
	return StartInExecutionContext[A](globalUnboundedExecutionContext)(io)
}

// FireAndForget starts the given IO and closes observation of its result.
// The underlying work continues independently; FireAndForget does not cancel it.
func FireAndForget[A any](ioa IO[A]) IO[fun.Unit] {
	return FlatMap(Start(ioa), func(fiber Fiber[A]) IO[fun.Unit] {
		return fiber.Close()
	})
}

type failedFiberImpl[A any] struct {
	Error error
}

// FailedFiber creates a fiber that will fail on Join or Close with the given error.
func FailedFiber[A any](err error) Fiber[A] {
	return &failedFiberImpl[A]{
		Error: err,
	}
}

func (f *failedFiberImpl[A]) Join() IO[A] {
	return Fail[A](f.Error)
}

func (f *failedFiberImpl[A]) Close() IO[fun.Unit] {
	return Fail[fun.Unit](f.Error)
}

// JoinFiberAsGoResult joins the fiber synchronously and returns GoResult.
func JoinFiberAsGoResult[A any](f Fiber[A]) GoResult[A] {
	return RunSync(f.Join())
}

// JoinWithTimeout joins the given fiber and waits no more than the given duration.
func JoinWithTimeout[A any](f Fiber[A], d time.Duration) IO[A] {
	return WithTimeout[A](d)(f.Join())
}
