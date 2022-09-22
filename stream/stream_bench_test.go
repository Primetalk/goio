package stream_test

import (
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/slice"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func BenchmarkStreamSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sumIO := stream.Head(stream.Sum(stream.Take(nats, 10000)))
		res, err1 := io.UnsafeRunSync(sumIO)
		assert.NoError(b, err1)
		assert.Equal(b, 50005000, res)
	}
}

var range10000 = func() (res []int) {
	res, _ = io.UnsafeRunSync(stream.ToSlice(stream.Take(nats, 10000)))
	return
}()

func BenchmarkSliceSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res := slice.Sum(range10000)
		assert.Equal(b, 50005000, res)
	}
}

func BenchmarkForSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 1; j <= 10000; j++ {
			sum += j
		}
		assert.Equal(b, 50005000, sum)
	}
}
