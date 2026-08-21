package io_test

import (
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireRecoveredPanic(t *testing.T, err error, panicMessage string) {
	t.Helper()

	require.Error(t, err)
	assert.Contains(t, err.Error(), panicMessage)
}

func TestLiftFuncApplicationIsLazy(t *testing.T) {
	applications := 0
	lifted := io.LiftFunc(func(value int) int {
		applications++
		return value + 1
	})
	require.Zero(t, applications)

	program := lifted(41)
	require.Zero(t, applications)

	result, err := io.UnsafeRunSync(program)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 1, applications)
}

func TestLiftFuncRunBoundaryRecoversApplicationPanic(t *testing.T) {
	executed := false
	lifted := io.LiftFunc(func(int) int {
		executed = true
		panic("LiftFunc application panic")
	})

	program := lifted(1)
	require.False(t, executed)

	_, err := io.UnsafeRunSync(program)

	assert.True(t, executed)
	requireRecoveredPanic(t, err, "LiftFunc application panic")
}

func TestEvalIsLazyAndRunBoundaryRecoversPanic(t *testing.T) {
	executed := false
	program := io.Eval(func() (int, error) {
		executed = true
		panic("Eval panic")
	})
	require.False(t, executed)

	_, err := io.UnsafeRunSync(program)

	assert.True(t, executed)
	requireRecoveredPanic(t, err, "Eval panic")
}

func TestDelayIsLazyAndRunBoundaryRecoversPanic(t *testing.T) {
	executed := false
	program := io.Delay(func() io.IO[int] {
		executed = true
		panic("Delay panic")
	})
	require.False(t, executed)

	_, err := io.UnsafeRunSync(program)

	assert.True(t, executed)
	requireRecoveredPanic(t, err, "Delay panic")
}

func TestMapIsLazyAndRunBoundaryRecoversMappingPanic(t *testing.T) {
	executed := false
	program := io.Map(io.Lift(1), func(int) int {
		executed = true
		panic("Map panic")
	})
	require.False(t, executed)

	_, err := io.UnsafeRunSync(program)

	assert.True(t, executed)
	requireRecoveredPanic(t, err, "Map panic")
}

func TestFlatMapIsLazyAndRunBoundaryRecoversBindingPanic(t *testing.T) {
	executed := false
	program := io.FlatMap(io.Lift(1), func(int) io.IO[int] {
		executed = true
		panic("FlatMap panic")
	})
	require.False(t, executed)

	_, err := io.UnsafeRunSync(program)

	assert.True(t, executed)
	requireRecoveredPanic(t, err, "FlatMap panic")
}

func TestAsyncRegistrationIsLazyAndRunBoundaryRecoversPanic(t *testing.T) {
	executed := false
	program := io.Async[int](func(io.Callback[int]) {
		executed = true
		panic("Async registration panic")
	})
	require.False(t, executed)

	result := runAsyncWithTimeout(t, program)

	assert.True(t, executed)
	requireRecoveredPanic(t, result.Error, "Async registration panic")
}

func TestStartedFiberPublishesRecoveredWorkPanic(t *testing.T) {
	work := io.Eval(func() (int, error) {
		panic("started fiber panic")
	})
	fiber, err := io.UnsafeRunSync(io.Start(work))
	require.NoError(t, err)

	result := runAsyncWithTimeout(t, fiber.Join())

	requireRecoveredPanic(t, result.Error, "started fiber panic")
}

func TestDirectIOInvocationRemainsOrdinaryGoCall(t *testing.T) {
	program := io.Eval(func() (int, error) {
		panic("direct invocation panic")
	})

	assert.PanicsWithValue(t, "direct invocation panic", func() {
		_ = program()
	})
}
