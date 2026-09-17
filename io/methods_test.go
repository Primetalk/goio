package io_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/require"
)

func TestGenericMethods(t *testing.T) {
	program := io.Lift(42).
		Map(strconv.Itoa).
		FlatMap(func(value string) io.IO[int] {
			return io.Lift(len(value))
		}).
		Fold(
			func(length int) io.IO[string] { return io.Lift(strconv.Itoa(length)) },
			func(err error) io.IO[string] { return io.Lift(err.Error()) },
		)

	result, err := io.UnsafeRunSync(program)
	require.NoError(t, err)
	require.Equal(t, "2", result)
}

func TestGenericFoldMethodHandlesFailure(t *testing.T) {
	expected := errors.New("boom")
	program := io.Fail[int](expected).Fold(
		func(value int) io.IO[string] { return io.Lift(strconv.Itoa(value)) },
		func(err error) io.IO[string] { return io.Lift(err.Error()) },
	)

	result, err := io.UnsafeRunSync(program)
	require.NoError(t, err)
	require.Equal(t, expected.Error(), result)
}
