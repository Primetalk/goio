package stream_test

import (
	"testing"

	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestTakeWhile(t *testing.T) {
	nats1112 := stream.TakeWhile(
		stream.DropWhile(
			nats,
			func(i int) bool { return i < 10 },
		),
		func(i int) bool { return i < 12 },
	)
	res := UnsafeStreamToSlice(t, nats1112)
	assert.ElementsMatch(t, res, []int{10, 11})
}
