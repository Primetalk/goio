package io

import "sync"

// Callback[A] is a function that takes A and error. A is only valid if error is nil.
type Callback[A any] func(A, error)

// Async[A] constructs an IO given a function that will eventually call a callback.
// Internally this function blocks the executing goroutine until the callback is called.
// Only the first callback invocation is observed; later invocations return without blocking.
// Async does not provide a cancellation token.
func Async[A any](k func(Callback[A])) IO[A] {
	return func() ResultOrContinuation[A] {
		ch := make(chan ResultOrContinuation[A], 1)
		var once sync.Once
		cb := func(a A, err error) {
			once.Do(func() {
				ch <- ResultOrContinuation[A]{
					Value: a,
					Error: err,
				}
			})
		}
		k(cb)
		res := <-ch
		return res
	}
}

// StartInGoRoutineAndWaitForResult - not very useful function.
// While it executes the IO in the go routine, the current
// thread is blocked.
func StartInGoRoutineAndWaitForResult[A any](io IO[A]) IO[A] {
	goRoutine := func(cb Callback[A]) {
		a, err1 := UnsafeRunSync(io)
		cb(a, err1)
	}
	return Async(func(cb Callback[A]) {
		go goRoutine(cb)
	})
}
