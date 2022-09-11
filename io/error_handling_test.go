package io_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

func TestRetryS(t *testing.T) {
	i := -6
	incFail := io.FromUnit(func() error {
		i += 1
		if i >= 0 {
			return nil
		} else {
			return errExpected
		}
	})
	retried2 := io.RetryS(incFail, io.RetryStrategyMaxCount("expected"), 2)
	UnsafeIOExpectError(t, errExpected, retried2)
	retried3 := io.RetryS(incFail, io.RetryStrategyMaxCount("expected"), 3)
	assert.Equal(t, fun.NewPair(fun.Unit1, 1), UnsafeIO(t, retried3))
}

func TestRetry(t *testing.T) {
	i := -6
	incFail := io.FromUnit(func() error {
		i += 1
		if i >= 0 {
			return nil
		} else {
			return errExpected
		}
	})
	retried2 := io.Retry(incFail, io.RetryStrategyMaxCount("expected"), 2)
	UnsafeIOExpectError(t, errExpected, retried2)
	retried3 := io.Retry(incFail, io.RetryStrategyMaxCount("expected"), 3)
	assert.Equal(t, fun.Unit1, UnsafeIO(t, retried3))
}
