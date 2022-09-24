package io

// Callback[A] is a function that takes A and error. A is only valid if error is nil.
type Callback[A any] func(A, error)

// Async[A] constructs an IO given a function that will eventually call a callback.
// Internally this function creates a channel and blocks on it until the function calls it.
func Async[A any](k func(Callback[A])) IO[A] {
	return func() ResultOrContinuation[A] {
		ch := make(chan ResultOrContinuation[A])
		cb := func(a A, err error) {
			ch <- ResultOrContinuation[A]{
				Value: a,
				Error: err,
			}
			close(ch)
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
