package stream_test

import (
	"testing"

	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestConcatPipes(t *testing.T) {
	inc := stream.MapPipe(func(i int) int { return i + 1 })
	dec := stream.MapPipe(func(i int) int { return i - 1 })
	nop := stream.ConcatPipes(inc, dec)
	assert.ElementsMatch(t, nats10Values, UnsafeStreamToSlice(t, nop(nats10)))
}
