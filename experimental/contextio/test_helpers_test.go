package contextio_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/primetalk/goio/experimental/contextio"
)

type intResult struct {
	value int
	err   error
}

type atomicInt32 struct {
	value int32
}

func (a *atomicInt32) Add(delta int32) int32 {
	return atomic.AddInt32(&a.value, delta)
}

func (a *atomicInt32) Load() int32 {
	return atomic.LoadInt32(&a.value)
}

type atomicBool struct {
	value uint32
}

func (a *atomicBool) Store(value bool) {
	var stored uint32
	if value {
		stored = 1
	}
	atomic.StoreUint32(&a.value, stored)
}

func (a *atomicBool) Load() bool {
	return atomic.LoadUint32(&a.value) != 0
}

func runIntAsync(effect contextio.Effect[int]) <-chan intResult {
	return runIntAsyncContext(context.Background(), effect)
}

func runIntAsyncContext(ctx context.Context, effect contextio.Effect[int]) <-chan intResult {
	result := make(chan intResult, 1)
	go func() {
		value, err := contextio.Run(ctx, effect)
		result <- intResult{value: value, err: err}
	}()
	return result
}

func receive[T any](t *testing.T, values <-chan T) T {
	t.Helper()
	select {
	case value := <-values:
		return value
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for value")
		var zero T
		return zero
	}
}

func receiveSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	receive(t, signal)
}

func assertNoResult[T any](t *testing.T, values <-chan T) {
	t.Helper()
	select {
	case <-values:
		t.Fatal("received a result before manual release")
	case <-time.After(20 * time.Millisecond):
	}
}

func assertPanicError(t *testing.T, err error) {
	t.Helper()
	var panicErr *contextio.PanicError
	if !errors.As(err, &panicErr) {
		t.Fatalf("error = %T %v, want PanicError", err, err)
	}
}
