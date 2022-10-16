package stream_test

import (
	"testing"

	"github.com/primetalk/goio/slice"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestZipWithIndex(t *testing.T) {
	natsWithIndex := UnsafeStreamToSlice(t, stream.ZipWithIndex(nats10))
	sNatsWithIndex := slice.ZipWithIndex(slice.Nats(10))
	assert.ElementsMatch(t, sNatsWithIndex, natsWithIndex)
}
