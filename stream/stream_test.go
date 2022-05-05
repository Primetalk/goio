package stream_test

import (
	"fmt"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	empty := stream.Empty[int]()
	_, err := io.UnsafeRunSync(stream.DrainAll(empty))
	assert.Equal(t, nil, err)
	stream10_12 := stream.LiftMany(10, 11, 12)
	stream20_24 := stream.Map(stream10_12, func(i int)int { return i * 2 })
	res, err := io.UnsafeRunSync(stream.ToSlice(stream20_24))
	assert.Equal(t, nil, err)
	assert.Equal(t, []int{20,22,24}, res)
}

var printInt = stream.NewSink(func (i int) { fmt.Printf("%d", i)})
func TestGenerate(t *testing.T) {
	powers2 := stream.Unfold(1, func(s int) int{
		return s * 2
	})

	res, err := io.UnsafeRunSync(stream.Head(powers2))
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, res)

	powers2_10 := stream.Drop(powers2, 9)
	res, err = io.UnsafeRunSync(stream.Head(powers2_10))
	assert.Equal(t, nil, err)
	assert.Equal(t, 1024, res)
}
