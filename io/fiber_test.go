package io_test

import (
	"testing"
	"time"

	"github.com/primetalk/goio/io"
)

func TestJoinWithTimeout(t *testing.T) {
	helloAfterSleeping100ms := io.AfterTimeout(100*time.Millisecond, io.Lift("hello"))
	fibIO := io.Start(helloAfterSleeping100ms)
	UnsafeIOExpectError(t, io.ErrorTimeout, io.FlatMap(fibIO, func(fib io.Fiber[string]) io.IO[string] {
		return io.JoinWithTimeout(fib, 10*time.Millisecond)
	}))
}
