package io_test

import (
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

func TestIO(t *testing.T) {
	io10 := io.Lift(10)
	io20 := io.Map(io10, func(i int)int { return i * 2 })
	io30 := io.FlatMap(io10, func(i int)io.IO[int]{ 
		return io.MapErr(io20, func(j int)(int, error){
			return i + j, nil
		})
	})
	res, err := io.UnsafeRunSync(io30)
	assert.Equal(t, err, nil)
	assert.Equal(t, res, 30)
}
