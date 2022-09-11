package stream_test

import (
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/slice"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func BenchmarkStreamSum(b *testing.B) {
	sumIO := stream.Head(stream.Sum(stream.Take(nats, 10000)))
	res, err1 := io.UnsafeRunSync(sumIO)
	assert.NoError(b, err1)
	assert.Equal(b, 50005000, res)
}

var range10000 = func() (res []int) {
	res, _ = io.UnsafeRunSync(stream.ToSlice(stream.Take(nats, 10000)))
	return
}()

func BenchmarkSliceSum(b *testing.B) {
	res := slice.Sum(range10000)
	assert.Equal(b, 50005000, res)
}
