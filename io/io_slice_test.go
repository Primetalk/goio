package io_test

import (
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/slice"
	"github.com/stretchr/testify/assert"
)

func TestMapSlice(t *testing.T) {
	assert.ElementsMatch(t,
		UnsafeIO(t, io.MapSlice(io.Lift(slice.Nats(5)), func(i int) int { return -i })),
		[]int{-1, -2, -3, -4, -5})
}
