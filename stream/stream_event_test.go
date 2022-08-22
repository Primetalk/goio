package stream_test

import (
	"testing"

	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestNormalFinish(t *testing.T) {
	se := UnsafeIO(t, stream.Last(stream.ToStreamEvent(nats10)))
	assert.Equal(t, stream.StreamEvent[int]{IsFinished: true}, se)
}


func TestStreamEventOfFailedStream(t *testing.T) {
	se := UnsafeIO(t, stream.Last(stream.ToStreamEvent(failedStream)))
	assert.Equal(t, stream.StreamEvent[int]{Error: errExpected}, se)
}
