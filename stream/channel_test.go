package stream_test

import (
	"testing"

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
