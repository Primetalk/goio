package io

import (
	"github.com/primetalk/goio/fun"
)

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


type fiberImpl[A any] struct {
	callbacks chan Callback[A]
}

func (f fiberImpl[A])Join() IO[A] {
	return Async(func(cb Callback[A]){
		f.callbacks <- cb
	})
}

func (f fiberImpl[A])Close() IO[fun.Unit] {
	return FromUnit(func() error{
		close(f.callbacks)
		return nil
	})
}

var maxCallbackCount = 16

// Start will start the IO in a separate go-routine.
// It'll establish a channel with callbacks, so that
// any number of listeners could join the returned fiber.
// When completed it'll start sending the results to the callbacks.
// The same value will be delivered to all listeners.
func Start[A any](io IO[A]) IO[Fiber[A]] {
	return Pure(func()(Fiber[A]){
		callbacks := make(chan Callback[A], maxCallbackCount)
		goRoutine := func(){
			a, err1 := UnsafeRunSync(io)
			for cb := range callbacks {
				cb(a, err1)
			}
		}
		go goRoutine()
		fiber := fiberImpl[A]{
			callbacks: callbacks,
		}
		return fiber
	})
}

// FireAndForget runs the given IO in a go routine and ignores the result
// It uses Fiber underneath.
func FireAndForget[A any](ioa IO[A]) IO[fun.Unit] {
	return FlatMap(Start(ioa), func (fiber Fiber[A])IO[fun.Unit] { 
		return fiber.Close()
	})
}
