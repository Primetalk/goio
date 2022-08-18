package stream_test

import (
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestChunksResize(t *testing.T) {
	chunks3 := stream.ToChunks[int](3)(nats10)
	chunks5 := stream.ChunksResize[int](4)(chunks3)
	res, err1 := io.UnsafeRunSync(stream.ToSlice(chunks5))
	assert.NoError(t, err1)
	assert.ElementsMatch(t, [][]int{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10}}, res)
}
