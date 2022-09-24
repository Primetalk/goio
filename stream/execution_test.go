package stream_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
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

func TestPartition(t *testing.T) {
	cdIO := stream.Partition(nats10, isEven,
		func(even stream.Stream[int]) io.IO[int] {
			return stream.Head(stream.Sum(even))
		},
		func(odd stream.Stream[int]) io.IO[string] {
			return stream.Head(stream.Map(stream.Sum(odd), fun.ToString[int]))
		},
	)
	res, err := io.UnsafeRunSync(cdIO)
	assert.NoError(t, err)
	assert.Equal(t, fun.NewPair(30, "25"), res)
}

func TestHeadAndTail(t *testing.T) {
	p := UnsafeIO(t, stream.HeadAndTail(nats))
	assert.Equal(t, 1, p.V1)
	stm2 := p.V2
	assert.Equal(t, 2, UnsafeIO(t, stream.Head(stm2)))
}

func TestTakeAndTail(t *testing.T) {
	p := UnsafeIO(t, stream.TakeAndTail(nats, 5, []int{}))
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5}, p.V1)
	stm2 := p.V2
	assert.Equal(t, 6, UnsafeIO(t, stream.Head(stm2)))
}
