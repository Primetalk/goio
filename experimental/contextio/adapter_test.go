package contextio_test

import (
	"context"
	"errors"
	"testing"

	"github.com/primetalk/goio/experimental/contextio"
	currentio "github.com/primetalk/goio/io"
)

func TestAdaptersAreLazyAndPropagateResults(t *testing.T) {
	var fromCalls atomicInt32
	from := contextio.FromIO(currentio.Eval(func() (int, error) {
		fromCalls.Add(1)
		return 42, nil
	}))
	if fromCalls.Load() != 0 {
		t.Fatal("FromIO executed eagerly")
	}
	value, err := contextio.Run(context.Background(), from)
	if err != nil || value != 42 || fromCalls.Load() != 1 {
		t.Fatalf("FromIO = (%d, %v), calls %d", value, err, fromCalls.Load())
	}

	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, 7)
	var toCalls atomicInt32
	to := contextio.ToIO(ctx, contextio.Eval(func(ctx context.Context) (int, error) {
		toCalls.Add(1)
		return ctx.Value(key{}).(int), nil
	}))
	if toCalls.Load() != 0 {
		t.Fatal("ToIO executed eagerly")
	}
	value, err = currentio.UnsafeRunSync(to)
	if err != nil || value != 7 || toCalls.Load() != 1 {
		t.Fatalf("ToIO = (%d, %v), calls %d", value, err, toCalls.Load())
	}
}

func TestAdaptersPropagateErrorsAndPanics(t *testing.T) {
	wantErr := errors.New("adapter")
	value, err := contextio.Run(context.Background(), contextio.FromIO(currentio.Fail[int](wantErr)))
	if value != 0 || !errors.Is(err, wantErr) {
		t.Fatalf("FromIO failure = (%d, %v)", value, err)
	}

	_, err = contextio.Run(context.Background(), contextio.FromIO(currentio.Eval(func() (int, error) {
		panic("current IO")
	})))
	if err == nil {
		t.Fatal("FromIO panic was not converted to an error")
	}

	_, err = currentio.UnsafeRunSync(contextio.ToIO(context.Background(), contextio.Eval(func(context.Context) (int, error) {
		panic("context effect")
	})))
	assertPanicError(t, err)
}

func TestFromIONonCooperativeCancellationWaitsForCompletion(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	current := currentio.Eval(func() (int, error) {
		close(started)
		<-release
		return 42, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	result := runIntAsyncContext(ctx, contextio.FromIO(current))
	receiveSignal(t, started)
	cancel()
	assertNoResult(t, result)
	close(release)
	got := receive(t, result)
	if got.value != 0 || !errors.Is(got.err, context.Canceled) {
		t.Fatalf("canceled FromIO = (%d, %v)", got.value, got.err)
	}
}
