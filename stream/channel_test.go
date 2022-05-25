package stream_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestSendingDataThroughChannel(t *testing.T) {
	ch := make(chan int)
	pipe := stream.PairOfChannelsToPipe(ch, ch)
	nats10After := stream.Through(nats10, pipe)
	results, err := io.UnsafeRunSync(stream.ToSlice(nats10After))
	assert.NoError(t, err)
	assert.ElementsMatch(t, results, nats10Values)
}

func TestStreamConversion(t *testing.T) {
	io2 := io.ForEach(pipeMul2IO, func (pair fun.Pair[chan int, chan int]) {
		input := pair.V1
		output := pair.V2
		input <- 10
		o := <-output
		assert.Equal(t, 20, o)
	})
	_, err := io.UnsafeRunSync(io2)
	assert.NoError(t, err)
}
