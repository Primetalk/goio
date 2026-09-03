// Package contextio provides an experimental context-aware effect prototype.
//
// The package is intended for architecture evaluation only. It provides
// cooperative cancellation: user work that ignores its context can delay
// cancellation and structured cleanup indefinitely.
package contextio

import "context"

// Unit is the zero-sized result of effects that produce no meaningful value.
type Unit struct{}

// Effect is an opaque, lazy context-aware computation.
// Its zero value is invalid and Run returns ErrInvalidEffect for it.
type Effect[A any] struct {
	run func(context.Context) (A, error)
}

// Succeed constructs an effect that succeeds with value.
func Succeed[A any](value A) Effect[A] {
	return Effect[A]{run: func(context.Context) (A, error) {
		return value, nil
	}}
}

// Fail constructs an effect that fails with err.
func Fail[A any](err error) Effect[A] {
	return Effect[A]{run: func(context.Context) (A, error) {
		var zero A
		return zero, err
	}}
}

// Eval constructs a lazy effect from context-aware synchronous work.
func Eval[A any](f func(context.Context) (A, error)) Effect[A] {
	return Effect[A]{run: func(ctx context.Context) (A, error) {
		return f(ctx)
	}}
}

// Defer lazily constructs the effect to execute.
func Defer[A any](f func() Effect[A]) Effect[A] {
	return Effect[A]{run: func(ctx context.Context) (A, error) {
		if err := ctx.Err(); err != nil {
			var zero A
			return zero, err
		}
		return runEffect(ctx, f())
	}}
}

// Map transforms a successful effect value.
func Map[A, B any](effect Effect[A], f func(A) B) Effect[B] {
	return Effect[B]{run: func(ctx context.Context) (B, error) {
		value, err := runEffect(ctx, effect)
		if err != nil {
			var zero B
			return zero, err
		}
		if err := ctx.Err(); err != nil {
			var zero B
			return zero, err
		}
		return f(value), nil
	}}
}

// MapErr transforms a successful effect value with a function that may fail.
func MapErr[A, B any](effect Effect[A], f func(A) (B, error)) Effect[B] {
	return Effect[B]{run: func(ctx context.Context) (B, error) {
		value, err := runEffect(ctx, effect)
		if err != nil {
			var zero B
			return zero, err
		}
		if err := ctx.Err(); err != nil {
			var zero B
			return zero, err
		}
		return f(value)
	}}
}

// FlatMap constructs and executes the next effect after a success.
func FlatMap[A, B any](effect Effect[A], f func(A) Effect[B]) Effect[B] {
	return Effect[B]{run: func(ctx context.Context) (B, error) {
		value, err := runEffect(ctx, effect)
		if err != nil {
			var zero B
			return zero, err
		}
		if err := ctx.Err(); err != nil {
			var zero B
			return zero, err
		}
		return runEffect(ctx, f(value))
	}}
}

// Fold handles either the successful value or the failure with another effect.
func Fold[A, B any](effect Effect[A], success func(A) Effect[B], failure func(error) Effect[B]) Effect[B] {
	return Effect[B]{run: func(ctx context.Context) (B, error) {
		value, err := runEffect(ctx, effect)
		if ctxErr := ctx.Err(); ctxErr != nil {
			var zero B
			return zero, ctxErr
		}
		if err != nil {
			return runEffect(ctx, failure(err))
		}
		return runEffect(ctx, success(value))
	}}
}

// Run executes an effect synchronously using ctx.
func Run[A any](ctx context.Context, effect Effect[A]) (A, error) {
	return runEffect(ctx, effect)
}

func runEffect[A any](ctx context.Context, effect Effect[A]) (value A, err error) {
	if ctx == nil {
		return value, ErrNilContext
	}
	if effect.run == nil {
		return value, ErrInvalidEffect
	}
	if err = ctx.Err(); err != nil {
		return value, err
	}

	completed := false
	defer func() {
		if !completed {
			value = *new(A)
			err = &PanicError{Value: recover()}
			return
		}
		if err != nil {
			value = *new(A)
		}
	}()

	value, err = effect.run(ctx)
	completed = true
	return value, err
}

func callEffect[A, B any](f func(A) Effect[B], value A) (effect Effect[B], err error) {
	completed := false
	defer func() {
		if !completed {
			effect = Effect[B]{}
			err = &PanicError{Value: recover()}
		}
	}()
	effect = f(value)
	completed = true
	return effect, nil
}
