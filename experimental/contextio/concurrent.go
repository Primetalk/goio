package contextio

import (
	"context"
	"time"
)

// Sleep waits for duration or cooperative context cancellation.
func Sleep(duration time.Duration) Effect[Unit] {
	return Effect[Unit]{run: func(ctx context.Context) (Unit, error) {
		timer := time.NewTimer(duration)
		defer timer.Stop()
		select {
		case <-timer.C:
			return Unit{}, nil
		case <-ctx.Done():
			return Unit{}, ctx.Err()
		}
	}}
}

// Timeout runs effect with a deadline, cancels it when the deadline wins, and
// waits for terminal child publication before returning. Context-ignoring work
// can therefore delay timeout completion indefinitely.
func Timeout[A any](duration time.Duration, effect Effect[A]) Effect[A] {
	return Effect[A]{run: func(ctx context.Context) (A, error) {
		timeoutCtx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()

		fiber := startFiber(timeoutCtx, effect, nil)
		select {
		case <-fiber.done:
			return fiber.readResult()
		default:
		}

		select {
		case <-fiber.done:
			return fiber.readResult()
		case <-timeoutCtx.Done():
			fiber.signalCancel()
			fiber.await()
			var zero A
			return zero, timeoutCtx.Err()
		}
	}}
}

type indexedOutcome[A any] struct {
	index  int
	result outcome[A]
}

// Race returns the first observed terminal result, whether success or failure.
// It cancels and joins every loser before returning. A context-ignoring loser can
// therefore delay completion indefinitely.
func Race[A any](effects []Effect[A]) Effect[A] {
	return Effect[A]{run: func(ctx context.Context) (A, error) {
		if len(effects) == 0 {
			var zero A
			return zero, ErrEmptyRace
		}

		results := make(chan indexedOutcome[A], len(effects))
		fibers := make([]*fiberState[A], 0, len(effects))
		for index, effect := range effects {
			index := index
			fiber := startFiber(ctx, effect, func(result outcome[A]) {
				results <- indexedOutcome[A]{index: index, result: result}
			})
			fibers = append(fibers, fiber)
		}

		var winner indexedOutcome[A]
		select {
		case winner = <-results:
		default:
			select {
			case winner = <-results:
			case <-ctx.Done():
				cancelAndJoinAll(fibers, -1)
				var zero A
				return zero, ctx.Err()
			}
		}

		cancelAndJoinAll(fibers, winner.index)
		return winner.result.value, winner.result.err
	}}
}

func cancelAndJoinAll[A any](fibers []*fiberState[A], winner int) {
	for index, fiber := range fibers {
		if index != winner {
			fiber.signalCancel()
		}
	}
	for index, fiber := range fibers {
		if index != winner {
			fiber.await()
		}
	}
}
