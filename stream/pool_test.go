package stream_test

import (
	"errors"
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

	sleepResults := stream.ThroughPipeEval(sleepTasks100, poolIO)
	resultStream := stream.MapEval(sleepResults, io.FromConstantGoResult[int])
	resultsIO := stream.ToSlice(resultStream)

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
	taskCount := 100

	sleepTasks100 := stream.Take(sleepTasks, taskCount)
	concurrency := 10
	ec := io.BoundedExecutionContext(concurrency, 0)
	sleepResults := stream.ThroughExecutionContext(sleepTasks100, ec, concurrency)
	resultsIO := stream.ToSlice(sleepResults)

	start := time.Now()
	results, err := io.UnsafeRunSync(resultsIO)
	assert.NoError(t, err)
	assert.Equal(t, taskCount, slice.SetSize(slice.ToSet(results)))
	required_duration := 10*taskCount/concurrency + 50
	assert.WithinDuration(t, start, time.Now(), time.Duration(required_duration)*time.Millisecond)
}

func TestFailedStreamThroughEC(t *testing.T) {
	expectedError := errors.New("expected error")
	failedStream := stream.Eval(io.Fail[io.IO[int]](expectedError))
	bec := io.BoundedExecutionContext(10, 0)
	throu := stream.ThroughExecutionContext(failedStream, bec, 10)
	_, err1 := io.UnsafeRunSync(stream.DrainAll(throu))
	if assert.Error(t, err1) {
		assert.Equal(t, expectedError, err1)
	}
}

func TestFailedDataStreamThroughEC(t *testing.T) {
	expectedError := errors.New("expected error")
	failedStream := stream.Lift(io.Fail[int](expectedError))
	bec := io.BoundedExecutionContext(10, 0)
	throu := stream.ThroughExecutionContext(failedStream, bec, 10)
	_, err1 := io.UnsafeRunSync(stream.DrainAll(throu))
	if assert.Error(t, err1) {
		assert.Equal(t, expectedError, err1)
	}
}
