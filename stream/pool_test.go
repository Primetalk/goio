package stream_test

import (
	"testing"
	"time"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/slice"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	sleep10ms := func(id int) io.IO[int] {
		return io.Pure(func() int {
			time.Sleep(10 * time.Millisecond)
			return id
		})
	}
	sleepTasks := stream.Map(nats, sleep10ms)
	sleepTasks100 := stream.Take(sleepTasks, 100)

	poolIO := stream.NewPool[int](10)

	resultsIO := io.FlatMap(poolIO, func(pool stream.Pool[int]) io.IO[[]int] {
		sleepResults := stream.ThroughPool(sleepTasks100, pool)
		resultStream := stream.MapEval(sleepResults, io.FromConstantGoResult[int])
		return stream.ToSlice(resultStream)
	})
	start := time.Now()
	results, err := io.UnsafeRunSync(resultsIO)
	assert.NoError(t, err)
	assert.Equal(t, 100, slice.SetSize(slice.ToSet(results)))
	assert.WithinDuration(t, start, time.Now(), 200*time.Millisecond)
}

func TestExecutionContext(t *testing.T) {
	sleep10ms := func(id int) io.IO[int] {
		return io.Pure(func() int {
			time.Sleep(10 * time.Millisecond)
			return id
		})
	}
	sleepTasks := stream.Map(nats, sleep10ms)
	sleepTasks100 := stream.Take(sleepTasks, 100)

	concurrency := 2
	ec := io.BoundedExecutionContext(concurrency,0)
	// poolIO := stream.NewPoolFromExecutionContext[int](ec, concurrency)

	// resultsIO := io.FlatMap(poolIO, func(pool stream.Pool[int]) io.IO[[]int] {
	sleepResults := stream.ThroughExecutionContext(sleepTasks100, ec, concurrency)
	resultStream := stream.MapEval(sleepResults, io.FromConstantGoResult[int])
	resultsIO := stream.ToSlice(resultStream)

	start := time.Now()
	results, err := io.UnsafeRunSync(resultsIO)
	assert.NoError(t, err)
	assert.Equal(t, 100, slice.SetSize(slice.ToSet(results)))
	assert.WithinDuration(t, start, time.Now(), 200*time.Millisecond)
}
