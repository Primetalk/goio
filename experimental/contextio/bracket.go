package contextio

import (
	"context"
	"errors"
	"time"
)

// ExitCase identifies how bracket use terminated.
type ExitCase uint8

const (
	// ExitSucceeded indicates successful use.
	ExitSucceeded ExitCase = iota
	// ExitErrored indicates failed or panicking use.
	ExitErrored
	// ExitCanceled indicates use terminated by context cancellation or deadline.
	ExitCanceled
)

// Bracket acquires a resource, uses it, and releases it exactly once after
// successful acquisition. Release executes with cancellation masked while
// retaining parent context values.
func Bracket[R, A any](
	acquire Effect[R],
	use func(R) Effect[A],
	release func(R, ExitCase) Effect[Unit],
) Effect[A] {
	return Effect[A]{run: func(ctx context.Context) (A, error) {
		resource, acquireErr := runEffect(ctx, acquire)
		if acquireErr != nil {
			var zero A
			return zero, acquireErr
		}

		var value A
		useErr := ctx.Err()
		if useErr == nil {
			var useEffect Effect[A]
			useEffect, useErr = callEffect(use, resource)
			if useErr == nil {
				value, useErr = runEffect(ctx, useEffect)
				if useErr == nil {
					useErr = ctx.Err()
				}
			}
		}

		exitCase := classifyExit(useErr)
		releaseEffect, releaseConstructionErr := callRelease(release, resource, exitCase)
		releaseErr := releaseConstructionErr
		if releaseErr == nil {
			_, releaseErr = runEffect(uncancelableContext{parent: ctx}, releaseEffect)
		}

		if useErr != nil {
			value = *new(A)
			if releaseErr != nil {
				return value, &CombinedError{Primary: useErr, Secondary: releaseErr}
			}
			return value, useErr
		}
		if releaseErr != nil {
			return *new(A), releaseErr
		}
		return value, nil
	}}
}

func classifyExit(err error) ExitCase {
	if err == nil {
		return ExitSucceeded
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ExitCanceled
	}
	return ExitErrored
}

func callRelease[R any](release func(R, ExitCase) Effect[Unit], resource R, exitCase ExitCase) (effect Effect[Unit], err error) {
	completed := false
	defer func() {
		if !completed {
			effect = Effect[Unit]{}
			err = &PanicError{Value: recover()}
		}
	}()
	effect = release(resource, exitCase)
	completed = true
	return effect, nil
}

type uncancelableContext struct {
	parent context.Context
}

func (ctx uncancelableContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (ctx uncancelableContext) Done() <-chan struct{} {
	return nil
}

func (ctx uncancelableContext) Err() error {
	return nil
}

func (ctx uncancelableContext) Value(key any) any {
	return ctx.parent.Value(key)
}
