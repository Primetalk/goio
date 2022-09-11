package io_test

import (
	"errors"
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected error")

func UnsafeIO[A any](t *testing.T, ioa io.IO[A]) A {
	res, err1 := io.UnsafeRunSync(ioa)
	assert.NoError(t, err1)
	return res
}

func UnsafeIOExpectError[A any](t *testing.T, expected error, ioa io.IO[A]) {
	_, err1 := io.UnsafeRunSync(ioa)
	if assert.Error(t, err1) {
		assert.Equal(t, expected, err1)
	}
}

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
