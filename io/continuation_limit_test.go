package io

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withContinuationLimit(t *testing.T, limit int) {
	t.Helper()

	previous := MaxContinuationDepth
	MaxContinuationDepth = limit
	t.Cleanup(func() {
		MaxContinuationDepth = previous
	})
}

func twoStepContinuation(value int) Continuation[int] {
	final := Continuation[int](func() ResultOrContinuation[int] {
		return ResultOrContinuation[int]{Value: value}
	})
	return func() ResultOrContinuation[int] {
		return ResultOrContinuation[int]{Continuation: &final}
	}
}

func TestObtainResultSucceedsAtExactContinuationLimit(t *testing.T) {
	withContinuationLimit(t, 2)

	result, err := ObtainResult(twoStepContinuation(42))

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestObtainResultFailsWhenContinuationLimitIsExceeded(t *testing.T) {
	withContinuationLimit(t, 1)

	_, err := ObtainResult(twoStepContinuation(42))

	require.Error(t, err)
	assert.Equal(t, "couldn't enforce continuation in 1 iterations", err.Error())
}

func TestObtainResultRejectsNilInitialContinuation(t *testing.T) {
	withContinuationLimit(t, 1)

	_, err := ObtainResult[int](nil)

	require.Error(t, err)
	assert.Equal(t, "nil continuation is being enforced", err.Error())
}

func TestObtainResultRejectsNilIntermediateContinuation(t *testing.T) {
	withContinuationLimit(t, 2)
	var nilContinuation Continuation[int]
	initial := Continuation[int](func() ResultOrContinuation[int] {
		return ResultOrContinuation[int]{Continuation: &nilContinuation}
	})

	_, err := ObtainResult(initial)

	require.Error(t, err)
	assert.Equal(t, "nil continuation is being enforced", err.Error())
}

func TestObtainResultZeroLimitDoesNotInvokeContinuation(t *testing.T) {
	withContinuationLimit(t, 0)
	invocations := 0
	continuation := Continuation[int](func() ResultOrContinuation[int] {
		invocations++
		return ResultOrContinuation[int]{Value: 42}
	})

	_, err := ObtainResult(continuation)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "0 iterations")
	assert.Zero(t, invocations)
}

func TestObtainResultNegativeLimitDoesNotInvokeContinuation(t *testing.T) {
	withContinuationLimit(t, -1)
	invocations := 0
	continuation := Continuation[int](func() ResultOrContinuation[int] {
		invocations++
		return ResultOrContinuation[int]{Value: 42}
	})

	_, err := ObtainResult(continuation)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "-1 iterations")
	assert.Zero(t, invocations)
}

func TestObtainResultSnapshotsContinuationLimit(t *testing.T) {
	withContinuationLimit(t, 2)
	final := Continuation[int](func() ResultOrContinuation[int] {
		return ResultOrContinuation[int]{Value: 42}
	})
	initial := Continuation[int](func() ResultOrContinuation[int] {
		MaxContinuationDepth = 0
		return ResultOrContinuation[int]{Continuation: &final}
	})

	result, err := ObtainResult(initial)

	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Zero(t, MaxContinuationDepth)
}

func TestObtainResultLimitErrorUsesSnapshot(t *testing.T) {
	withContinuationLimit(t, 1)
	next := Continuation[int](func() ResultOrContinuation[int] {
		return ResultOrContinuation[int]{Value: 42}
	})
	initial := Continuation[int](func() ResultOrContinuation[int] {
		MaxContinuationDepth = 99
		return ResultOrContinuation[int]{Continuation: &next}
	})

	_, err := ObtainResult(initial)

	require.Error(t, err)
	assert.True(t, strings.HasSuffix(err.Error(), "1 iterations"), err.Error())
}
