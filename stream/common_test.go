package stream_test

import (
	"fmt"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

var nats = stream.Unfold(0, func(s int) int {
	return s + 1
})
var nats10 = stream.Take(nats, 10)
var nats10Values = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

var Mul2 = stream.MapPipe(func(i int) int { return i * 2 })
var pipeMul2IO = stream.PipeToPairOfChannels(Mul2)

var printInt = stream.NewSink(func(i int) { fmt.Printf("%d", i) })

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
