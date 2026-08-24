package contextio_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/primetalk/goio/experimental/contextio"
)

func TestAsyncSynchronousAndAsynchronousCompletion(t *testing.T) {
	var registered atomicBool
	synchronous := contextio.Async(func(_ context.Context, callback contextio.Callback[int]) contextio.Canceler {
		registered.Store(true)
		callback(42, nil)
		return nil
	})
	if registered.Load() {
		t.Fatal("Async registered during construction")
	}
	value, err := contextio.Run(context.Background(), synchronous)
	if err != nil || value != 42 {
		t.Fatalf("synchronous Async = (%d, %v)", value, err)
	}

	callbackReady := make(chan contextio.Callback[int], 1)
	asynchronous := contextio.Async(func(_ context.Context, callback contextio.Callback[int]) contextio.Canceler {
		callbackReady <- callback
		return nil
	})
	result := runIntAsync(asynchronous)
	callback := receive(t, callbackReady)
	callback(7, nil)
	got := receive(t, result)
	if got.value != 7 || got.err != nil {
		t.Fatalf("asynchronous Async = (%d, %v)", got.value, got.err)
	}
}

func TestAsyncFirstCallbackWins(t *testing.T) {
	effect := contextio.Async(func(_ context.Context, callback contextio.Callback[int]) contextio.Canceler {
		var wait sync.WaitGroup
		wait.Add(20)
		for value := 0; value < 20; value++ {
			value := value
			go func() {
				defer wait.Done()
				callback(value, nil)
			}()
		}
		wait.Wait()
		callback(99, nil)
		return nil
	})
	value, err := contextio.Run(context.Background(), effect)
	if err != nil || value < 0 || value >= 20 {
		t.Fatalf("Async duplicate result = (%d, %v)", value, err)
	}
}

func TestAsyncCancellationCallsCancelerOnceAndWaits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callbackReady := make(chan contextio.Callback[int], 1)
	canceled := make(chan struct{})
	var cancelCalls atomicInt32
	effect := contextio.Async(func(_ context.Context, callback contextio.Callback[int]) contextio.Canceler {
		callbackReady <- callback
		return func() {
			if cancelCalls.Add(1) == 1 {
				close(canceled)
			}
		}
	})

	result := runIntAsyncContext(ctx, effect)
	callback := receive(t, callbackReady)
	cancel()
	receiveSignal(t, canceled)
	assertNoResult(t, result)
	callback(42, nil)
	got := receive(t, result)
	if got.value != 0 || !errors.Is(got.err, context.Canceled) || cancelCalls.Load() != 1 {
		t.Fatalf("canceled Async = (%d, %v), cancel calls %d", got.value, got.err, cancelCalls.Load())
	}
}

func TestAsyncRegistrationAndCancelerPanics(t *testing.T) {
	_, err := contextio.Run(context.Background(), contextio.Async[int](func(context.Context, contextio.Callback[int]) contextio.Canceler {
		panic("register")
	}))
	assertPanicError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	callbackReady := make(chan contextio.Callback[int], 1)
	cancelerCalled := make(chan struct{})
	effect := contextio.Async(func(_ context.Context, callback contextio.Callback[int]) contextio.Canceler {
		callbackReady <- callback
		return func() {
			close(cancelerCalled)
			panic("cancel")
		}
	})
	result := runIntAsyncContext(ctx, effect)
	callback := receive(t, callbackReady)
	cancel()
	receiveSignal(t, cancelerCalled)
	callback(0, context.Canceled)
	got := receive(t, result)
	assertPanicError(t, got.err)
}

func TestAsyncNonCooperativeCancellationNeedsManualRelease(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callbackReady := make(chan contextio.Callback[int], 1)
	effect := contextio.Async(func(_ context.Context, callback contextio.Callback[int]) contextio.Canceler {
		callbackReady <- callback
		return nil
	})
	result := runIntAsyncContext(ctx, effect)
	callback := receive(t, callbackReady)
	cancel()
	assertNoResult(t, result)
	callback(1, nil)
	got := receive(t, result)
	if !errors.Is(got.err, context.Canceled) {
		t.Fatalf("non-cooperative cancellation error = %v", got.err)
	}
}

func TestAsyncCallbackCancellationRaceHasOneTerminalOutcome(t *testing.T) {
	for iteration := 0; iteration < 100; iteration++ {
		ctx, cancel := context.WithCancel(context.Background())
		callbackReady := make(chan contextio.Callback[int], 1)
		effect := contextio.Async(func(_ context.Context, callback contextio.Callback[int]) contextio.Canceler {
			callbackReady <- callback
			return nil
		})
		result := runIntAsyncContext(ctx, effect)
		callback := receive(t, callbackReady)
		var racers sync.WaitGroup
		racers.Add(2)
		go func() {
			defer racers.Done()
			cancel()
		}()
		go func() {
			defer racers.Done()
			callback(42, nil)
		}()
		racers.Wait()
		got := receive(t, result)
		if got.err == nil {
			if got.value != 42 {
				t.Fatalf("successful race value = %d", got.value)
			}
		} else if got.value != 0 || !errors.Is(got.err, context.Canceled) {
			t.Fatalf("canceled race = (%d, %v)", got.value, got.err)
		}
	}
}
