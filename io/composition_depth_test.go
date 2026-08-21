package io_test

import (
	"errors"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const supportedCompositionDepth = 10_000

func TestFlatMapSupportsBoundedLeftAssociatedDepth(t *testing.T) {
	program := io.Lift(0)
	for step := 0; step < supportedCompositionDepth; step++ {
		program = io.FlatMap(program, func(value int) io.IO[int] {
			return io.Lift(value + 1)
		})
	}

	result, err := io.UnsafeRunSync(program)

	require.NoError(t, err)
	assert.Equal(t, supportedCompositionDepth, result)
}

func TestMapSupportsBoundedLeftAssociatedDepth(t *testing.T) {
	program := io.Lift(0)
	for step := 0; step < supportedCompositionDepth; step++ {
		program = io.Map(program, func(value int) int {
			return value + 1
		})
	}

	result, err := io.UnsafeRunSync(program)

	require.NoError(t, err)
	assert.Equal(t, supportedCompositionDepth, result)
}

func TestSequenceSupportsBoundedDepth(t *testing.T) {
	programs := make([]io.IO[int], supportedCompositionDepth)
	for index := range programs {
		programs[index] = io.Lift(index)
	}

	result, err := io.UnsafeRunSync(io.Sequence(programs))

	require.NoError(t, err)
	require.Len(t, result, supportedCompositionDepth)
	for index, value := range result {
		assert.Equal(t, index, value)
	}
}

func TestMapPropagatesFailureAtBoundedDepth(t *testing.T) {
	expectedErr := errors.New("deep composition failed")
	mappingCalls := 0
	program := io.Fail[int](expectedErr)
	for step := 0; step < supportedCompositionDepth; step++ {
		program = io.Map(program, func(value int) int {
			mappingCalls++
			return value + 1
		})
	}

	_, err := io.UnsafeRunSync(program)

	require.ErrorIs(t, err, expectedErr)
	assert.Zero(t, mappingCalls)
}
