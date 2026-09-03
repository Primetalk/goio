package contextio

import (
	"context"
	"sync"
)

// Callback publishes an asynchronous terminal result. Only its first call has
// an effect.
type Callback[A any] func(A, error)

// Canceler synchronously requests cancellation of registered asynchronous work.
type Canceler func()

type asyncResult[A any] struct {
	value A
	err   error
}

func noOp() {}

// Async constructs an effect from prompt one-shot callback registration.
// If cancellation wins, the canceler is invoked once and execution waits for
// terminal callback publication. Non-cooperative registration or work can
// therefore delay cancellation indefinitely.
func Async[A any](register func(context.Context, Callback[A]) Canceler) Effect[A] {
	return Effect[A]{run: func(ctx context.Context) (A, error) {
		results := make(chan asyncResult[A], 1)
		var once sync.Once
		callback := Callback[A](func(value A, err error) {
			once.Do(func() {
				if err != nil {
					value = *new(A)
				}
				results <- asyncResult[A]{value: value, err: err}
			})
		})

		canceler, err := callRegister(ctx, register, callback)
		if err != nil {
			var zero A
			return zero, err
		}
		if canceler == nil {
			canceler = noOp
		}

		// Give an already-published synchronous callback deterministic precedence.
		select {
		case result := <-results:
			return result.value, result.err
		default:
		}

		select {
		case result := <-results:
			return result.value, result.err
		case <-ctx.Done():
			cancelErr := callCanceler(canceler)
			<-results // wait for strict terminal completion even if the canceler panics
			var zero A
			if cancelErr != nil {
				return zero, cancelErr
			}
			return zero, ctx.Err()
		}
	}}
}

// calls user provided register function and catches panic.
func callRegister[A any](ctx context.Context, register func(context.Context, Callback[A]) Canceler, callback Callback[A]) (canceler Canceler, err error) {
	completed := false
	defer func() {
		if !completed {
			canceler = nil
			err = &PanicError{Value: recover()}
		}
	}()
	canceler = register(ctx, callback)
	completed = true
	return canceler, nil
}

// calls user provided canceler function and catches panic.
func callCanceler(canceler Canceler) (err error) {
	completed := false
	defer func() {
		if !completed {
			err = &PanicError{Value: recover()}
		}
	}()
	canceler()
	completed = true
	return nil
}
