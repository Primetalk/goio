package resource_test

import (
	"testing"
	"time"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/resource"
	"github.com/stretchr/testify/assert"
)

func CreateSleeps(count int) (ios []io.IO[int]) {
	sleep100ms := io.SleepA(100*time.Millisecond, "a")
	for i := 0; i < count; i += 1 {
		ios = append(ios, io.Map(sleep100ms, fun.Const[string](i)))
	}
	return
}

func TestParallelBound(t *testing.T) {
	start := time.Now()
	becRes := resource.BoundedExecutionContextResource(50, 0)
	ioall := resource.Use(becRes, func(bec io.ExecutionContext) io.IO[[]int] {
		return io.ParallelInExecutionContext[int](bec)(CreateSleeps(100))
	})
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
