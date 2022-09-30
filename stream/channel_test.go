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

func TestToChannels(t *testing.T) {
	ch1 := make(chan int, 10)
	ch2 := make(chan int, 10)
	drainIO := stream.ToChannels(nats10, ch1, ch2)
	UnsafeIO(t, drainIO)
	slice1IO := stream.ToSlice(stream.FromChannel(ch1))
	slice1 := UnsafeIO(t, slice1IO)
	assert.ElementsMatch(t, nats10Values, slice1)
	slice2IO := stream.ToSlice(stream.FromChannel(ch2))
	slice2 := UnsafeIO(t, slice2IO)
	assert.ElementsMatch(t, nats10Values, slice2)
}
func TestStreamConversion(t *testing.T) {
	io2 := io.ForEach(pipeMul2IO, func(pair fun.Pair[chan<- int, <-chan int]) {
		input := pair.V1
		output := pair.V2
		input <- 10
		o := <-output
		assert.Equal(t, 20, o)
	})
	UnsafeIO(t, io2)
}
