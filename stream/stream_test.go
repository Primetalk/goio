package stream_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	empty := stream.Empty[int]()
	_, err := io.UnsafeRunSync(stream.DrainAll(empty))
	assert.Equal(t, nil, err)
	stream10_12 := stream.LiftMany(10, 11, 12)
	stream20_24 := Mul2(stream10_12)
	res, err := io.UnsafeRunSync(stream.ToSlice(stream20_24))
	assert.Equal(t, nil, err)
	assert.Equal(t, []int{20, 22, 24}, res)
}

func TestGenerate(t *testing.T) {
	powers2 := stream.Unfold(1, func(s int) int {
		return s * 2
	})

	res, err := io.UnsafeRunSync(stream.Head(powers2))
	assert.NoError(t, err)
	assert.Equal(t, 2, res)

	powers2_10 := stream.Drop(powers2, 9)
	res, err = io.UnsafeRunSync(stream.Head(powers2_10))
	assert.NoError(t, err)
	assert.Equal(t, 1024, res)
}

func TestDrainAll(t *testing.T) {
	results := []int{}
	natsAppend := stream.MapEval(
		stream.Take(stream.Repeat(nats10), 10),
		func(a int) io.IO[int] {
			return io.Eval(func() (int, error) {
				results = append(results, a)
				return a, nil
			})
		})
	_, err := io.UnsafeRunSync(stream.DrainAll(natsAppend))
	assert.NoError(t, err)
	assert.ElementsMatch(t, results, nats10Values)
}

func TestStateFlatMap(t *testing.T) {
	sumStream := stream.Sum(nats10)
	ioSum := stream.Head(sumStream)
	sum, err := io.UnsafeRunSync(ioSum)
	assert.NoError(t, err)
	assert.Equal(t, 55, sum)
}

func isEven(i int) bool {
	return i%2 == 0
}

func TestFlatMapPipe(t *testing.T) {
	natsRepeated := stream.FlatMapPipe(func(i int) stream.Stream[int] {
		return stream.MapPipe(func(j int) int {
			return i + j
		})(nats10)
	})(nats10)

	ioLen := stream.Head(stream.Len(natsRepeated))
	len, err := io.UnsafeRunSync(ioLen)
	assert.NoError(t, err)
	assert.Equal(t, 100, len)

	filtered := stream.Filter(natsRepeated, isEven)
	sumStream := stream.Sum(filtered)
	ioSum := stream.Head(sumStream)
	var sum int
	sum, err = io.UnsafeRunSync(ioSum)
	assert.NoError(t, err)
	assert.Equal(t, 550, sum)
}

func TestChunks(t *testing.T) {
	natsBy10 := stream.ChunkN[int](10)(stream.Take(nats, 19))
	nats10to19IO := stream.Head(stream.Drop(natsBy10, 1))
	nats10to19, err := io.UnsafeRunSync(nats10to19IO)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []int{11, 12, 13, 14, 15, 16, 17, 18, 19}, nats10to19)
}

func TestGroupBy(t *testing.T) {
	groupedNats := stream.GroupBy(stream.Take(nats, 7), func(i int) int {
		return i / 5
	})
	groupsIO := stream.ToSlice(groupedNats)
	groups, err := io.UnsafeRunSync(groupsIO)
	assert.NoError(t, err)
	expected := []fun.Pair[int, []int]{
		{V1: 0, V2: []int{1, 2, 3, 4}},
		{V1: 1, V2: []int{5, 6, 7}},
	}
	assert.ElementsMatch(t, expected, groups)
}
