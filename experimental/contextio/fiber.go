package contextio

import (
	"context"
	"sync"
)

// Fiber is a running prototype effect with explicit cancellation and joining.
type Fiber[A any] interface {
	Join() Effect[A]
	Cancel() Effect[Unit]
}

type outcome[A any] struct {
	value A
	err   error
}

type fiberState[A any] struct {
	mu        sync.Mutex
	done      chan struct{}
	ctx       context.Context
	cancel    context.CancelFunc
	completed bool
	cancelWon bool
	result    outcome[A]
}

// Start lazily starts effect in exactly one work goroutine under a derived child
// context.
func Start[A any](effect Effect[A]) Effect[Fiber[A]] {
	return Effect[Fiber[A]]{run: func(ctx context.Context) (Fiber[A], error) {
		return startFiber(ctx, effect, nil), nil
	}}
}

func startFiber[A any](ctx context.Context, effect Effect[A], notify func(outcome[A])) *fiberState[A] {
	childCtx, cancel := context.WithCancel(ctx)
	state := &fiberState[A]{
		done:   make(chan struct{}),
		ctx:    childCtx,
		cancel: cancel,
	}
	go func() {
		value, err := Run(childCtx, effect)
		published := state.publish(outcome[A]{value: value, err: err})
		if notify != nil {
			notify(published)
		}
	}()
	return state
}

func (f *fiberState[A]) publish(result outcome[A]) outcome[A] {
	f.mu.Lock()
	if f.cancelWon {
		result.value = *new(A)
		result.err = f.ctx.Err()
		if result.err == nil {
			result.err = context.Canceled
		}
	} else if err := f.ctx.Err(); err != nil {
		result.value = *new(A)
		result.err = err
	}
	if result.err != nil {
		result.value = *new(A)
	}
	f.result = result
	f.completed = true
	close(f.done)
	f.mu.Unlock()
	f.cancel()
	return result
}

// Join waits for the immutable terminal fiber outcome. Canceling only the Join
// execution context stops that observation; it does not cancel the fiber.
func (f *fiberState[A]) Join() Effect[A] {
	return Effect[A]{run: func(ctx context.Context) (A, error) {
		select {
		case <-f.done:
			return f.readResult()
		default:
		}
		select {
		case <-f.done:
			return f.readResult()
		case <-ctx.Done():
			var zero A
			return zero, ctx.Err()
		}
	}}
}

// Cancel idempotently requests child cancellation and waits for terminal
// publication. Non-cooperative work can delay completion indefinitely.
func (f *fiberState[A]) Cancel() Effect[Unit] {
	return Effect[Unit]{run: func(context.Context) (Unit, error) {
		f.signalCancel()
		<-f.done
		return Unit{}, nil
	}}
}

func (f *fiberState[A]) signalCancel() {
	f.mu.Lock()
	won := !f.completed
	if won {
		f.cancelWon = true
	}
	f.mu.Unlock()
	if won {
		f.cancel()
	}
}

func (f *fiberState[A]) readResult() (A, error) {
	<-f.done
	f.mu.Lock()
	result := f.result
	f.mu.Unlock()
	return result.value, result.err
}

func (f *fiberState[A]) await() outcome[A] {
	value, err := f.readResult()
	return outcome[A]{value: value, err: err}
}
