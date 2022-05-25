package io_test

import (
	"errors"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

func TestIO(t *testing.T) {
	io10 := io.Lift(10)
	io20 := io.Map(io10, func(i int) int { return i * 2 })
	io30 := io.FlatMap(io10, func(i int) io.IO[int] {
		return io.MapErr(io20, func(j int) (int, error) {
			return i + j, nil
		})
	})
	res, err := io.UnsafeRunSync(io30)
	assert.Equal(t, err, nil)
	assert.Equal(t, res, 30)
}

func TestErr(t *testing.T) {
	var ptr *string = nil
	ptrio := io.Lift(ptr)
	uptr := io.FlatMap(ptrio, io.Unptr[string])
	_, err := io.UnsafeRunSync(uptr)
	assert.Equal(t, io.ErrorNPE, err)
	wrappedUptr := io.Wrapf(uptr, "my message %d", 10)
	_, err = io.UnsafeRunSync(wrappedUptr)
	assert.Equal(t, "my message 10: nil pointer", err.Error())
}

func TestFinally(t *testing.T) {
	errorMessage := "on purpose failure"
	failure := io.Fail[string](errors.New(errorMessage))
	finalizerExecuted := false
	fin := io.Finally(failure, io.FromPureEffect(func() { finalizerExecuted = true }))
	_, err := io.UnsafeRunSync(fin)
	assert.Error(t, err, errorMessage)
	assert.True(t, finalizerExecuted)
}
