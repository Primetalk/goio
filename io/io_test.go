package io_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
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
	res := UnsafeIO(t, io30)
	assert.Equal(t, 30, res)
}

func TestLiftFunc(t *testing.T) {
	f := io.LiftFunc(inc)
	assert.Equal(t, 11, UnsafeIO(t, f(10)))
}

func TestErr(t *testing.T) {
	var ptr *string = nil
	ptrio := io.Lift(ptr)
	uptr := io.FlatMap(ptrio, io.Unptr[string])
	UnsafeIOExpectError(t, io.ErrorNPE, uptr)
	wrappedUptr := io.Wrapf(uptr, "my message %d", 10)
	_, err := io.UnsafeRunSync(wrappedUptr)
	assert.Equal(t, "my message 10: nil pointer", err.Error())
}

func TestFinally(t *testing.T) {
	finalizerExecuted := false
	onErrorExecuted := false
	fin := io.Finally(failure, io.FromPureEffect(func() { finalizerExecuted = true }))
	oe := io.OnError(fin, func(err error) io.IO[fun.Unit] {
		return io.FromPureEffect(func() { onErrorExecuted = true })
	})
	UnsafeIOExpectError(t, errExpected, oe)
	assert.True(t, finalizerExecuted)
	assert.True(t, onErrorExecuted)
}

func TestSequence(t *testing.T) {
	ios := Nats(10)
	for i, io1 := range ios {
		res := UnsafeIO(t, io1)
		assert.Equal(t, i, res)
	}
	ioseq := io.Sequence(ios)
	res, err := io.UnsafeRunSync(ioseq)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, res)
}

func TestIgnore(t *testing.T) {
	iou := io.Ignore(io.Lift(10))
	u, _ := io.UnsafeRunSync(iou)
	assert.Equal(t, fun.Unit1, u)
}
