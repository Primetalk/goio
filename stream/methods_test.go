package stream_test

import (
	"strconv"
	"testing"

	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/require"
)

func TestGenericMethods(t *testing.T) {
	stm := stream.LiftMany(1, 2).
		Map(strconv.Itoa).
		FlatMap(func(value string) stream.Stream[string] {
			return stream.LiftMany(value, value)
		}).
		Through(stream.MapPipe(func(value string) int { return len(value) }))

	result := UnsafeIO(t, stream.ToSlice(stm))
	require.Equal(t, []int{1, 1, 1, 1}, result)
}
