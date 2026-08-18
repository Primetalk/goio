package io

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fiberTestTimeout = time.Second

func newPendingFiber[A any]() *fiberImpl[A] {
	return &fiberImpl[A]{
		mu:        &sync.Mutex{},
		callbacks: []Callback[A]{},
	}
}

func runFiberJoin[A any](fiber Fiber[A]) <-chan GoResult[A] {
	result := make(chan GoResult[A], 1)
	go func() {
		result <- RunSync(fiber.Join())
	}()
	return result
}

func awaitFiberResult[A any](t *testing.T, result <-chan GoResult[A]) GoResult[A] {
	t.Helper()

	select {
	case res := <-result:
		return res
	case <-time.After(fiberTestTimeout):
		t.Fatal("timed out waiting for fiber result")
		return GoResult[A]{}
	}
}

func awaitRegisteredJoiners[A any](t *testing.T, fiber *fiberImpl[A], count int) {
	t.Helper()

	deadline := time.Now().Add(fiberTestTimeout)
	for {
		fiber.mu.Lock()
		registered := len(fiber.callbacks)
		fiber.mu.Unlock()
		if registered == count {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %d registered joiners; got %d", count, registered)
		}
		runtime.Gosched()
	}
}

func closeFiber[A any](t *testing.T, fiber Fiber[A]) {
	t.Helper()

	result := RunSync(fiber.Close())
	require.NoError(t, result.Error)
}

func assertFiberClosed[A any](t *testing.T, result GoResult[A]) {
	t.Helper()

	assert.ErrorIs(t, result.Error, ErrorFiberClosed)
}

func TestFiberCloseBeforeCompletionWithNoCurrentJoiners(t *testing.T) {
	fiber := newPendingFiber[int]()

	closeFiber(t, Fiber[int](fiber))

	assertFiberClosed(t, awaitFiberResult(t, runFiberJoin[int](Fiber[int](fiber))))
}

func TestFiberCloseWakesOneCurrentJoiner(t *testing.T) {
	fiber := newPendingFiber[int]()
	join := runFiberJoin[int](Fiber[int](fiber))
	awaitRegisteredJoiners(t, fiber, 1)

	closeFiber(t, Fiber[int](fiber))

	assertFiberClosed(t, awaitFiberResult(t, join))
	assertFiberClosed(t, awaitFiberResult(t, runFiberJoin[int](Fiber[int](fiber))))
}

func TestFiberCloseWakesMultipleCurrentJoiners(t *testing.T) {
	const joinerCount = 8

	fiber := newPendingFiber[int]()
	joins := make([]<-chan GoResult[int], 0, joinerCount)
	for i := 0; i < joinerCount; i++ {
		joins = append(joins, runFiberJoin[int](fiber))
	}
	awaitRegisteredJoiners(t, fiber, joinerCount)

	closeFiber[int](t, fiber)

	for _, join := range joins {
		assertFiberClosed(t, awaitFiberResult(t, join))
	}
	assertFiberClosed(t, awaitFiberResult(t, runFiberJoin[int](fiber)))
}

func TestFiberCloseIsIdempotent(t *testing.T) {
	fiber := newPendingFiber[int]()

	closeFiber[int](t, fiber)
	closeFiber[int](t, fiber)

	assertFiberClosed(t, awaitFiberResult(t, runFiberJoin[int](fiber)))
}

func TestFiberCompletionBeforeClosePreservesCompletedResult(t *testing.T) {
	fiber := newPendingFiber[int]()
	join := runFiberJoin[int](fiber)
	awaitRegisteredJoiners(t, fiber, 1)

	require.True(t, fiber.publishResult(GoResult[int]{Value: 42}))
	closeFiber[int](t, fiber)

	first := awaitFiberResult(t, join)
	require.NoError(t, first.Error)
	assert.Equal(t, 42, first.Value)
	late := awaitFiberResult(t, runFiberJoin[int](fiber))
	require.NoError(t, late.Error)
	assert.Equal(t, 42, late.Value)
}

func TestFiberCloseBeforeCompletionIgnoresLateResult(t *testing.T) {
	fiber := newPendingFiber[int]()
	join := runFiberJoin[int](fiber)
	awaitRegisteredJoiners(t, fiber, 1)

	closeFiber[int](t, fiber)
	require.False(t, fiber.publishResult(GoResult[int]{Value: 42}))

	assertFiberClosed(t, awaitFiberResult(t, join))
	assertFiberClosed(t, awaitFiberResult(t, runFiberJoin[int](fiber)))
}

func TestFiberCompletionAndCloseHaveOneTerminalWinner(t *testing.T) {
	testCases := []struct {
		name          string
		closeWins     bool
		expectedValue int
		expectedClose bool
	}{
		{name: "close wins", closeWins: true, expectedClose: true},
		{name: "completion wins", closeWins: false, expectedValue: 42},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fiber := newPendingFiber[int]()
			join := runFiberJoin[int](fiber)
			awaitRegisteredJoiners(t, fiber, 1)

			closeRelease := make(chan struct{})
			completionRelease := make(chan struct{})
			closeDone := make(chan struct{})
			completionDone := make(chan bool, 1)
			go func() {
				<-closeRelease
				RunSync(fiber.Close())
				close(closeDone)
			}()
			go func() {
				<-completionRelease
				completionDone <- fiber.publishResult(GoResult[int]{Value: 42})
			}()

			if testCase.closeWins {
				close(closeRelease)
				select {
				case <-closeDone:
				case <-time.After(fiberTestTimeout):
					t.Fatal("timed out waiting for close")
				}
				close(completionRelease)
				require.False(t, <-completionDone)
			} else {
				close(completionRelease)
				require.True(t, <-completionDone)
				close(closeRelease)
				select {
				case <-closeDone:
				case <-time.After(fiberTestTimeout):
					t.Fatal("timed out waiting for close")
				}
			}

			results := []GoResult[int]{
				awaitFiberResult(t, join),
				awaitFiberResult(t, runFiberJoin[int](fiber)),
			}
			for _, result := range results {
				if testCase.expectedClose {
					assertFiberClosed(t, result)
				} else {
					require.NoError(t, result.Error)
					assert.Equal(t, testCase.expectedValue, result.Value)
				}
			}
		})
	}
}

func TestFiberCallbacksRunOutsideMutex(t *testing.T) {
	fiber := newPendingFiber[int]()
	callbackDone := make(chan struct{})
	fiber.registerJoiner(func(int, error) {
		RunSync(fiber.Close())
		close(callbackDone)
	})

	require.True(t, fiber.publishResult(GoResult[int]{Value: 42}))

	select {
	case <-callbackDone:
	case <-time.After(fiberTestTimeout):
		t.Fatal("callback blocked while re-entering fiber")
	}
}

func TestFiberCloseDoesNotStopUnderlyingWork(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})
	work := Eval(func() (int, error) {
		close(started)
		<-release
		close(finished)
		return 42, nil
	})
	fiberResult := RunSync(Start(work))
	require.NoError(t, fiberResult.Error)
	fiber := fiberResult.Value

	select {
	case <-started:
	case <-time.After(fiberTestTimeout):
		t.Fatal("timed out waiting for fiber work to start")
	}
	closeFiber(t, fiber)
	close(release)
	select {
	case <-finished:
	case <-time.After(fiberTestTimeout):
		t.Fatal("underlying work did not continue after close")
	}

	assertFiberClosed(t, awaitFiberResult(t, runFiberJoin[int](fiber)))
}
