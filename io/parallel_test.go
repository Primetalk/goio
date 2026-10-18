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
	for _, first := range []string{"main", "other"} {
		t.Run(first+" finishes first", func(t *testing.T) {
			mainStarted, otherStarted := make(chan struct{}), make(chan struct{})
			mainRelease, otherRelease := make(chan struct{}), make(chan struct{})
			defer releaseSignal(mainRelease)
			defer releaseSignal(otherRelease)
			mainFinished, otherFinished := make(chan struct{}), make(chan struct{})
			main := controlledResult(mainStarted, mainRelease, mainFinished, "a", nil)
			other := controlledResult(otherStarted, otherRelease, otherFinished, fun.Unit1, nil)

			result := runIO(io.RunAlso(main, other))
			awaitSignal(t, mainStarted, "main effect to start")
			awaitSignal(t, otherStarted, "other effect to start concurrently")

			if first == "main" {
				releaseSignal(mainRelease)
				awaitSignal(t, mainFinished, "main effect to finish")
			} else {
				releaseSignal(otherRelease)
				awaitSignal(t, otherFinished, "other effect to finish")
			}
			select {
			case <-result:
				t.Fatal("RunAlso returned before both effects finished")
			default:
			}

			releaseSignal(mainRelease)
			releaseSignal(otherRelease)
			res := awaitRunResult(t, result, "both effects to complete")
			assert.NoError(t, res.Error)
			assert.Equal(t, "a", res.Value)
		})
	}
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
