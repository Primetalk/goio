package io_test

import (
	"errors"
	"log"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/slice"
	"github.com/stretchr/testify/assert"
)

var errorMessage = "expected error" //"on purpose failure"

var errExpected = errors.New(errorMessage)

var failure = io.Fail[string](errExpected)

func inc(i int) int {
	return i + 1
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

func Nats(count int) (ios []io.IO[int]) {
	return slice.Map(slice.Range(0, count), func(i int) io.IO[int] {
		return io.Pure(func() int {
			log.Printf("executing %v\n", i)
			return i
		})
	})
}
