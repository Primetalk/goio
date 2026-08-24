package contextio

import (
	"context"

	currentio "github.com/primetalk/goio/io"
)

// FromIO lazily adapts a current IO to an Effect. Current IO does not accept a
// context, so cancellation remains blocked until the wrapped IO returns.
func FromIO[A any](effect currentio.IO[A]) Effect[A] {
	return Effect[A]{run: func(ctx context.Context) (A, error) {
		value, err := currentio.UnsafeRunSync(effect)
		if err != nil {
			var zero A
			return zero, err
		}
		if err := ctx.Err(); err != nil {
			var zero A
			return zero, err
		}
		return value, nil
	}}
}

// ToIO lazily adapts an Effect to current IO using an explicit captured context.
func ToIO[A any](ctx context.Context, effect Effect[A]) currentio.IO[A] {
	return currentio.Eval(func() (A, error) {
		return Run(ctx, effect)
	})
}
