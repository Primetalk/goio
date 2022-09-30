package stream_test

import (
	"testing"
	"time"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/set"
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

	sleepResults := stream.ThroughPipeEval(sleepTasks100, poolIO)
	resultStream := stream.MapEval(sleepResults, io.FromConstantGoResult[int])
	resultsIO := stream.ToSlice(resultStream)

	start := time.Now()
	results := UnsafeIO(t, resultsIO)
	assert.Equal(t, 100, set.SetSize(slice.ToSet(results)))
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
	taskCount := 100

	sleepTasks100 := stream.Take(sleepTasks, taskCount)
	concurrency := 10
	ec := io.BoundedExecutionContext(concurrency, 0)
	sleepResults := stream.ThroughExecutionContext(sleepTasks100, ec, concurrency)
	resultsIO := stream.ToSlice(sleepResults)

	start := time.Now()
	results := UnsafeIO(t, resultsIO)
	assert.Equal(t, taskCount, set.SetSize(slice.ToSet(results)))
	required_duration := 10*taskCount/concurrency + 50
	assert.WithinDuration(t, start, time.Now(), time.Duration(required_duration)*time.Millisecond)
}

func TestFailedStreamThroughEC(t *testing.T) {
	failedStream := stream.Eval(io.Fail[io.IO[int]](errExpected))
	bec := io.BoundedExecutionContext(10, 0)
	throu := stream.ThroughExecutionContext(failedStream, bec, 10)
	UnsafeIOExpectError(t, errExpected, stream.DrainAll(throu))
}

func TestFailedDataStreamThroughEC(t *testing.T) {
	failedStream := stream.Lift(io.Fail[int](errExpected))
	bec := io.BoundedExecutionContext(10, 0)
	throu := stream.ThroughExecutionContext(failedStream, bec, 10)
	UnsafeIOExpectError(t, errExpected, stream.DrainAll(throu))
}

func TestThroughExecutionContextUnordered(t *testing.T) {
	durMs := 1
	sleeps := func(id int) io.IO[int] {
		return io.Pure(func() int {
			time.Sleep(time.Duration(durMs) * time.Millisecond)
			return id
		})
	}
	sleepTaskInfs := stream.Map(nats, sleeps)
	taskCount := 1000

	sleepTasks := stream.Take(sleepTaskInfs, taskCount)
	concurrency := 10
	ec := io.BoundedExecutionContext(concurrency, 0)
	sleepResults := stream.ThroughExecutionContextUnordered(sleepTasks, ec, concurrency)
	resultsIO := stream.ToSlice(sleepResults)

	start := time.Now()
	results := UnsafeIO(t, resultsIO)
	assert.Equal(t, taskCount, set.SetSize(slice.ToSet(results)))
	required_duration := durMs * taskCount / concurrency * 2
	assert.WithinDuration(t, start, time.Now(), time.Duration(required_duration)*time.Millisecond)
}

func TestFailedDataStreamThroughECUnord(t *testing.T) {
	failedStream := stream.Lift(io.Fail[int](errExpected))
	bec := io.BoundedExecutionContext(10, 0)
	throu := stream.ThroughExecutionContextUnordered(failedStream, bec, 10)
	UnsafeIOExpectError(t, errExpected, stream.DrainAll(throu))
}
