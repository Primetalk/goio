package stream_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

var Mul2 = stream.MapPipe(func(i int) int { return i * 2 })
var pipeMul2IO = stream.PipeToPairOfChannels(Mul2)

var printInt = stream.NewSink(func(i int) { fmt.Printf("%d", i) })
var errExpected = errors.New("expected error")
var failedStream = stream.Eval(io.Fail[int](errExpected))

var natsAndThenFail = stream.AndThen(nats10, failedStream)

func UnsafeStreamToSlice[A any](t *testing.T, stm stream.Stream[A]) []A {
	return UnsafeIO(t, stream.ToSlice(stm))
}

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
