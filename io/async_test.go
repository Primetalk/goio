package io_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const asyncTestTimeout = time.Second

func runAsyncWithTimeout[A any](t *testing.T, ioa io.IO[A]) io.GoResult[A] {
	t.Helper()

	result := make(chan io.GoResult[A], 1)
	go func() {
		value, err := io.UnsafeRunSync(ioa)
		result <- io.GoResult[A]{Value: value, Error: err}
	}()

	select {
	case res := <-result:
		return res
	case <-time.After(asyncTestTimeout):
		t.Fatal("timed out waiting for Async completion")
		return io.GoResult[A]{}
	}
}

func TestAsyncSynchronousCallback(t *testing.T) {
	ioa := io.Async(func(cb io.Callback[int]) {
		cb(42, nil)
	})

	res := runAsyncWithTimeout(t, ioa)
	require.NoError(t, res.Error)
	assert.Equal(t, 42, res.Value)
}

func TestAsyncAsynchronousCallback(t *testing.T) {
	registered := make(chan struct{})
	release := make(chan struct{})
	ioa := io.Async(func(cb io.Callback[int]) {
		go func() {
			close(registered)
			<-release
			cb(42, nil)
		}()
	})

	result := make(chan io.GoResult[int], 1)
	go func() {
		value, err := io.UnsafeRunSync(ioa)
		result <- io.GoResult[int]{Value: value, Error: err}
	}()

	select {
	case <-registered:
	case <-time.After(asyncTestTimeout):
		t.Fatal("timed out waiting for Async registration")
	}
	close(release)

	select {
	case res := <-result:
		require.NoError(t, res.Error)
		assert.Equal(t, 42, res.Value)
	case <-time.After(asyncTestTimeout):
		t.Fatal("timed out waiting for asynchronous callback")
	}
}

func TestAsyncSequentialDuplicateCallbacksUseFirstResult(t *testing.T) {
	secondReturned := false
	ioa := io.Async(func(cb io.Callback[int]) {
		cb(1, nil)
		cb(2, errors.New("ignored"))
		secondReturned = true
	})

	res := runAsyncWithTimeout(t, ioa)
	require.NoError(t, res.Error)
	assert.Equal(t, 1, res.Value)
	assert.True(t, secondReturned)
}

func TestAsyncConcurrentDuplicateCallbacksCompleteOnce(t *testing.T) {
	const callbackCount = 32

	callbacksReturned := make(chan struct{})
	ioa := io.Async(func(cb io.Callback[int]) {
		start := make(chan struct{})
		var callbacks sync.WaitGroup
		callbacks.Add(callbackCount)
		for value := 0; value < callbackCount; value++ {
			value := value
			go func() {
				defer callbacks.Done()
				<-start
				cb(value, nil)
			}()
		}
		close(start)
		go func() {
			callbacks.Wait()
			close(callbacksReturned)
		}()
	})

	res := runAsyncWithTimeout(t, ioa)
	require.NoError(t, res.Error)
	assert.GreaterOrEqual(t, res.Value, 0)
	assert.Less(t, res.Value, callbackCount)

	select {
	case <-callbacksReturned:
	case <-time.After(asyncTestTimeout):
		t.Fatal("duplicate callbacks did not return")
	}
}

func TestAsyncRegistrationPanicBecomesRunError(t *testing.T) {
	ioa := io.Async(func(io.Callback[int]) {
		panic("registration failed")
	})

	res := runAsyncWithTimeout(t, ioa)
	require.Error(t, res.Error)
	assert.Contains(t, res.Error.Error(), "registration failed")
}
