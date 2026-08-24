package contextio_test

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"

	"github.com/primetalk/goio/experimental/contextio"
)

func TestFiberStartIsLazyAndJoinIsRepeatable(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomicInt32
	start := contextio.Start(contextio.Eval(func(context.Context) (int, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-release
		return 42, nil
	}))
	assertNoSignal(t, started)
	fiber, err := contextio.Run(context.Background(), start)
	if err != nil {
		t.Fatal(err)
	}
	receiveSignal(t, started)

	const joiners = 8
	results := make(chan intResult, joiners)
	for index := 0; index < joiners; index++ {
		go func() {
			value, err := contextio.Run(context.Background(), fiber.Join())
			results <- intResult{value: value, err: err}
		}()
	}
	close(release)
	for index := 0; index < joiners; index++ {
		result := receive(t, results)
		if result.err != nil || result.value != 42 {
			t.Fatalf("Join = (%d, %v)", result.value, result.err)
		}
	}
	value, err := contextio.Run(context.Background(), fiber.Join())
	if err != nil || value != 42 || calls.Load() != 1 {
		t.Fatalf("future Join = (%d, %v), calls %d", value, err, calls.Load())
	}
}

func TestFiberPublishesFailureAndPanic(t *testing.T) {
	wantErr := errors.New("failure")
	failed, _ := contextio.Run(context.Background(), contextio.Start(contextio.Fail[int](wantErr)))
	value, err := contextio.Run(context.Background(), failed.Join())
	if value != 0 || !errors.Is(err, wantErr) {
		t.Fatalf("failed Join = (%d, %v)", value, err)
	}

	panicked, _ := contextio.Run(context.Background(), contextio.Start(contextio.Eval(func(context.Context) (int, error) {
		panic("fiber")
	})))
	_, err = contextio.Run(context.Background(), panicked.Join())
	assertPanicError(t, err)
}

func TestFiberCancelIsIdempotentAndWaitsForCleanup(t *testing.T) {
	started := make(chan struct{})
	allowCleanup := make(chan struct{})
	cleaned := make(chan struct{})
	effect := contextio.Eval(func(ctx context.Context) (int, error) {
		close(started)
		<-ctx.Done()
		<-allowCleanup
		close(cleaned)
		return 0, ctx.Err()
	})
	fiber, _ := contextio.Run(context.Background(), contextio.Start(effect))
	receiveSignal(t, started)

	const cancelers = 8
	results := make(chan error, cancelers)
	var wait sync.WaitGroup
	wait.Add(cancelers)
	for index := 0; index < cancelers; index++ {
		go func() {
			defer wait.Done()
			_, err := contextio.Run(context.Background(), fiber.Cancel())
			results <- err
		}()
	}
	assertNoSignal(t, cleaned)
	close(allowCleanup)
	wait.Wait()
	for index := 0; index < cancelers; index++ {
		if err := receive(t, results); err != nil {
			t.Fatalf("Cancel error = %v", err)
		}
	}
	value, err := contextio.Run(context.Background(), fiber.Join())
	if value != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("Join after Cancel = (%d, %v)", value, err)
	}
}

func TestFiberParentCancellationAndCompletedResult(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	fiber, _ := contextio.Run(parent, contextio.Start(contextio.Eval(func(ctx context.Context) (int, error) {
		close(started)
		<-ctx.Done()
		return 0, ctx.Err()
	})))
	receiveSignal(t, started)
	cancel()
	_, err := contextio.Run(context.Background(), fiber.Join())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("parent-canceled Join error = %v", err)
	}

	completed, _ := contextio.Run(context.Background(), contextio.Start(contextio.Succeed(7)))
	value, err := contextio.Run(context.Background(), completed.Join())
	if err != nil || value != 7 {
		t.Fatalf("completed Join = (%d, %v)", value, err)
	}
	if _, err := contextio.Run(context.Background(), completed.Cancel()); err != nil {
		t.Fatal(err)
	}
	value, err = contextio.Run(context.Background(), completed.Join())
	if err != nil || value != 7 {
		t.Fatalf("Join after late Cancel = (%d, %v)", value, err)
	}
}

func TestFiberNonCooperativeWorkDelaysCancelUntilManualRelease(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	fiber, err := contextio.Run(context.Background(), contextio.Start(contextio.Eval(func(context.Context) (int, error) {
		close(started)
		<-release
		return 42, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	receiveSignal(t, started)

	canceled := make(chan error, 1)
	go func() {
		_, err := contextio.Run(context.Background(), fiber.Cancel())
		canceled <- err
	}()
	assertNoResult(t, canceled)
	close(release)
	if err := receive(t, canceled); err != nil {
		t.Fatalf("Cancel error = %v", err)
	}
	value, err := contextio.Run(context.Background(), fiber.Join())
	if value != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("Join after non-cooperative cancel = (%d, %v)", value, err)
	}
}

func TestFiberCompletionCancelRacePublishesOneImmutableOutcome(t *testing.T) {
	for iteration := 0; iteration < 100; iteration++ {
		started := make(chan struct{})
		release := make(chan struct{})
		fiber, err := contextio.Run(context.Background(), contextio.Start(contextio.Eval(func(context.Context) (int, error) {
			close(started)
			<-release
			return 42, nil
		})))
		if err != nil {
			t.Fatal(err)
		}
		receiveSignal(t, started)

		cancelResult := make(chan error, 1)
		go func() {
			_, err := contextio.Run(context.Background(), fiber.Cancel())
			cancelResult <- err
		}()
		close(release)
		if err := receive(t, cancelResult); err != nil {
			t.Fatalf("Cancel error = %v", err)
		}

		firstValue, firstErr := contextio.Run(context.Background(), fiber.Join())
		secondValue, secondErr := contextio.Run(context.Background(), fiber.Join())
		if firstValue != secondValue || !sameTerminalError(firstErr, secondErr) {
			t.Fatalf("mutable outcome: first=(%d, %v), second=(%d, %v)", firstValue, firstErr, secondValue, secondErr)
		}
		if firstErr == nil && firstValue != 42 {
			t.Fatalf("successful outcome = %d, want 42", firstValue)
		}
		if firstErr != nil && !errors.Is(firstErr, context.Canceled) {
			t.Fatalf("race error = %v, want cancellation", firstErr)
		}
	}
}

func sameTerminalError(left, right error) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return errors.Is(left, context.Canceled) && errors.Is(right, context.Canceled)
}

func TestRepeatedCooperativeStartCancelCompletesWithoutGoroutineGrowth(t *testing.T) {
	baseline := runtime.NumGoroutine()
	for iteration := 0; iteration < 100; iteration++ {
		started := make(chan struct{})
		fiber, err := contextio.Run(context.Background(), contextio.Start(contextio.Eval(func(ctx context.Context) (int, error) {
			close(started)
			<-ctx.Done()
			return 0, ctx.Err()
		})))
		if err != nil {
			t.Fatal(err)
		}
		receiveSignal(t, started)
		if _, err := contextio.Run(context.Background(), fiber.Cancel()); err != nil {
			t.Fatal(err)
		}
	}
	runtime.GC()
	if got := runtime.NumGoroutine(); got > baseline+8 {
		t.Fatalf("goroutine count grew from %d to %d", baseline, got)
	}
}

func assertNoSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
		t.Fatal("unexpected signal")
	default:
	}
}
