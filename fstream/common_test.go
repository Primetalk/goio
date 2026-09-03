package fstream_test

import (
	"testing"

	"github.com/primetalk/goio/fstream"
	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

var nats = fstream.Nats()
var nats10 = fstream.Take(nats, 10)
var nats10Values = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

var Mul2 = fstream.MapPipe(func(i int) int { return i * 2 })

func UnsafeStreamToSlice[A any](t *testing.T, stm fstream.Stream[A]) []A {
	return UnsafeIO(t, fstream.ToSlice(stm))
}

func UnsafeIOStreamToSlice[A any](t *testing.T, stm fstream.IOStream[A]) []A {
	return UnsafeIO(t, io.FlatMap(io.IO[fstream.Stream[A]](stm), fstream.ToSlice[A]))
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

func TestNats(t *testing.T) {
	assert.ElementsMatch(t, nats10Values, UnsafeIOStreamToSlice(t, nats10))
}
