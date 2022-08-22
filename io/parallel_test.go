package io_test

import (
	"testing"
	"time"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

func CreateSleeps(count int) (ios []io.IO[int]) {
	sleep100ms := io.SleepA(100*time.Millisecond, "a")
	for i := 0; i < count; i += 1 {
		ios = append(ios, io.Map(sleep100ms, fun.Const[string](i)))
	}
	return
}

func TestParallel(t *testing.T) {
	ioall := io.Parallel(CreateSleeps(100)...)
	measured := io.MeasureDuration(ioall)
	results, err := io.UnsafeRunSync(measured)
	assert.Equal(t, err, nil)
	assert.Equal(t, 0, results.V1[0])
	duration := results.V2
	now := time.Now()
	assert.WithinDuration(t, now.Add(duration), now, 200*time.Millisecond)
}

func TestParallelBound(t *testing.T) {
	start := time.Now()
	bec := io.BoundedExecutionContext(50, 0)
	ioall := io.ParallelInExecutionContext[int](bec)(CreateSleeps(100))
	results, err := io.UnsafeRunSync(ioall)
	assert.Equal(t, err, nil)
	dur := time.Since(start)
	assert.Equal(t, 0, results[0])
	assert.Equal(t, 1, results[1])
	assert.Equal(t, 2, results[2])
	// it should take longer than 200 ms, but less than 10 seconds
	assert.GreaterOrEqual(t, dur, 200*time.Millisecond)
	assert.LessOrEqual(t, dur, 300*time.Millisecond)
}

func TestPairParallelAndRunAlso(t *testing.T) {
	ioA := io.RunAlso(
		io.SleepA(100*time.Millisecond, "a"),
		io.SleepA(100*time.Millisecond, fun.Unit1),
	)
	measured := io.MeasureDuration(ioA)
	results, err := io.UnsafeRunSync(measured)
	assert.Equal(t, err, nil)
	assert.Equal(t, "a", results.V1)
	duration := results.V2
	now := time.Now()
	assert.WithinDuration(t, now.Add(duration), now, 120*time.Millisecond)
}

func TestPairSequentially(t *testing.T) {
	ioA1 := io.PairSequentially(
		io.SleepA(100*time.Millisecond, "a"),
		io.SleepA(100*time.Millisecond, 1),
	)
	measured := io.MeasureDuration(ioA1)
	results, err := io.UnsafeRunSync(measured)
	assert.Equal(t, err, nil)
	assert.Equal(t, fun.NewPair("a", 1), results.V1)
	duration := results.V2
	assert.GreaterOrEqual(t, duration, 200*time.Millisecond)
}
