package io

import (
	"errors"
	"sync"
	"time"

	"github.com/primetalk/goio/fun"
)

// Fiber[A] is a type safe representation of Go routine.
// One might Join() and receive the result of the go routine.
// After Close() subsequent joins will fail.
type Fiber[A any] interface {
	// Join waits for results of the fiber.
	// When fiber completes, this IO will complete and return the result.
	// After this fiber is closed, all join IOs fail immediately.
	Join() IO[A]
	// Closes the fiber and stops sending callbacks.
	// After closing, the respective go routine may complete
	// This is not Cancel, it does not send any signals to the fiber.
	// The work will still be done.
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

func (f *fiberImpl[A]) Join() IO[A] {
	return Async(func(cb Callback[A]) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.result == nil {
			f.callbacks = append(f.callbacks, cb)
		} else {
			// we run external function in a go routine just to make sure we are not locked forever
			go cb(f.result.Value, f.result.Error)
		}
	})
}

func (f *fiberImpl[A]) Close() IO[fun.Unit] {
	return FromPureEffect(func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.result == nil {
			f.result = &GoResult[A]{
				Error: errors.New("fiber is closed"),
			}
		}
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
				fiber.mu.Lock()
				fiber.result = &GoResult[A]{a, err1}
				callbacks := fiber.callbacks
				fiber.callbacks = []Callback[A]{}
				fiber.mu.Unlock()
				for _, cb := range callbacks {
					cb(a, err1)
				}
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

// FireAndForget runs the given IO in a go routine and ignores the result
// It uses Fiber underneath.
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
