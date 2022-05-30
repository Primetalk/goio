package stream_test

import (
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestForEach(t *testing.T) {
	powers2 := stream.Unfold(1, func(s int) int {
		return s * 2
	})
	is := []int{}
	forEachIO := stream.ForEach(stream.Take(powers2, 5), func(i int) {
		is = append(is, i)
	})
	_, err := io.UnsafeRunSync(forEachIO)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []int{2, 4, 8, 16, 32}, is)
}
