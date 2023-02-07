package io_test

import (
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

func TestMemoize(t *testing.T) {
	counter := 0
	inc := func(i int) io.IO[int] {
		return io.Pure(func() int {
			counter += 1
			return i + 1
		})
	}
	incm := io.Memoize(inc)
	assert.Equal(t, 2, UnsafeIO(t, incm(1)))
	assert.Equal(t, 2, UnsafeIO(t, incm(1)))
	assert.Equal(t, 3, UnsafeIO(t, incm(2)))
	assert.Equal(t, 2, counter)
}
