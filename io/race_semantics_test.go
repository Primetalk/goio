package io_test

import (
	"errors"
	"testing"
	"time"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const raceSemanticsTestTimeout = 2 * time.Second

func awaitSignal(t *testing.T, signal <-chan struct{}, description string) {
	t.Helper()

	select {
	case <-signal:
	case <-time.After(raceSemanticsTestTimeout):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func awaitRunResult[A any](t *testing.T, result <-chan io.GoResult[A], description string) io.GoResult[A] {
	t.Helper()

	select {
	case res := <-result:
		return res
	case <-time.After(raceSemanticsTestTimeout):
		t.Fatalf("timed out waiting for %s", description)
		return io.GoResult[A]{}
	}
}

func runIO[A any](ioa io.IO[A]) <-chan io.GoResult[A] {
	result := make(chan io.GoResult[A], 1)
	go func() {
		result <- io.RunSync(ioa)
	}()
	return result
}

func releaseSignal(signal chan struct{}) {
	select {
	case <-signal:
	default:
		close(signal)
	}
}

func controlledResult[A any](started chan<- struct{}, release <-chan struct{}, finished chan<- struct{}, value A, err error) io.IO[A] {
	return io.Eval(func() (A, error) {
		close(started)
		<-release
		close(finished)
		return value, err
	})
}

func TestWithTimeoutReturnsWhileLosingWorkContinues(t *testing.T) {
	loserStarted := make(chan struct{})
	loserRelease := make(chan struct{})
	defer releaseSignal(loserRelease)
	loserFinished := make(chan struct{})
	loser := controlledResult(loserStarted, loserRelease, loserFinished, "late", nil)
	timed := io.WithTimeout[string](100 * time.Millisecond)(loser)

	result := runIO(timed)
	awaitSignal(t, loserStarted, "timed work to start")
	timedResult := awaitRunResult(t, result, "timeout result")
	require.ErrorIs(t, timedResult.Error, io.ErrorTimeout)

	select {
	case <-loserFinished:
		t.Fatal("losing work finished before it was released")
	default:
	}
	releaseSignal(loserRelease)
	awaitSignal(t, loserFinished, "losing work to continue after timeout")
}

func TestConcurrentlyFirstReturnsFirstSuccess(t *testing.T) {
	winnerStarted := make(chan struct{})
	winnerRelease := make(chan struct{})
	defer releaseSignal(winnerRelease)
	winnerFinished := make(chan struct{})
	loserStarted := make(chan struct{})
	loserRelease := make(chan struct{})
	defer releaseSignal(loserRelease)
	loserFinished := make(chan struct{})
	winner := controlledResult(winnerStarted, winnerRelease, winnerFinished, 42, nil)
	loser := controlledResult(loserStarted, loserRelease, loserFinished, 7, nil)

	result := runIO(io.ConcurrentlyFirst([]io.IO[int]{winner, loser}))
	awaitSignal(t, winnerStarted, "winning computation to start")
	awaitSignal(t, loserStarted, "losing computation to start")
	releaseSignal(winnerRelease)
	winnerResult := awaitRunResult(t, result, "first successful result")
	require.NoError(t, winnerResult.Error)
	assert.Equal(t, 42, winnerResult.Value)

	releaseSignal(loserRelease)
	awaitSignal(t, loserFinished, "losing computation to finish")
}

func TestConcurrentlyFirstReturnsFirstFailure(t *testing.T) {
	expectedErr := errors.New("first failure")
	winnerStarted := make(chan struct{})
	winnerRelease := make(chan struct{})
	defer releaseSignal(winnerRelease)
	winnerFinished := make(chan struct{})
	loserStarted := make(chan struct{})
	loserRelease := make(chan struct{})
	defer releaseSignal(loserRelease)
	loserFinished := make(chan struct{})
	winner := controlledResult(winnerStarted, winnerRelease, winnerFinished, 0, expectedErr)
	loser := controlledResult(loserStarted, loserRelease, loserFinished, 7, nil)

	result := runIO(io.ConcurrentlyFirst([]io.IO[int]{winner, loser}))
	awaitSignal(t, winnerStarted, "failing computation to start")
	awaitSignal(t, loserStarted, "losing computation to start")
	releaseSignal(winnerRelease)
	winnerResult := awaitRunResult(t, result, "first failed result")
	require.ErrorIs(t, winnerResult.Error, expectedErr)

	releaseSignal(loserRelease)
	awaitSignal(t, loserFinished, "losing computation to finish")
}

func TestConcurrentlyFirstLosersCanCompleteAfterWinnerReturns(t *testing.T) {
	const loserCount = 32

	winnerStarted := make(chan struct{})
	winnerRelease := make(chan struct{})
	defer releaseSignal(winnerRelease)
	winnerFinished := make(chan struct{})
	winner := controlledResult(winnerStarted, winnerRelease, winnerFinished, -1, nil)
	loserRelease := make(chan struct{})
	defer releaseSignal(loserRelease)
	loserStarted := make([]chan struct{}, loserCount)
	loserFinished := make([]chan struct{}, loserCount)
	competitors := make([]io.IO[int], 0, loserCount+1)
	competitors = append(competitors, winner)
	for index := 0; index < loserCount; index++ {
		loserStarted[index] = make(chan struct{})
		loserFinished[index] = make(chan struct{})
		competitors = append(competitors, controlledResult(
			loserStarted[index],
			loserRelease,
			loserFinished[index],
			index,
			nil,
		))
	}

	result := runIO(io.ConcurrentlyFirst(competitors))
	awaitSignal(t, winnerStarted, "winning computation to start")
	for _, started := range loserStarted {
		awaitSignal(t, started, "losing computation to start")
	}
	releaseSignal(winnerRelease)
	winnerResult := awaitRunResult(t, result, "winning result")
	require.NoError(t, winnerResult.Error)
	assert.Equal(t, -1, winnerResult.Value)

	releaseSignal(loserRelease)
	for _, finished := range loserFinished {
		awaitSignal(t, finished, "losing computation to finish after winner returned")
	}
}
