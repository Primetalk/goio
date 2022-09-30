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
	_ = UnsafeIO(t, stream.DrainAll(empty))
	stream10_12 := stream.LiftMany(10, 11, 12)
	stream20_24 := Mul2(stream10_12)
	res := UnsafeIO(t, stream.ToSlice(stream20_24))
	assert.Equal(t, []int{20, 22, 24}, res)
}

func TestGenerate(t *testing.T) {
	powers2 := stream.Unfold(1, func(s int) int {
		return s * 2
	})

	res := UnsafeIO(t, stream.Head(powers2))
	assert.Equal(t, 2, res)

	powers2_10 := stream.Drop(powers2, 9)
	res = UnsafeIO(t, stream.Head(powers2_10))
	assert.Equal(t, 1024, res)

	res = UnsafeIO(t, stream.Last(stream.Take(powers2, 10)))
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
	_ = UnsafeIO(t, stream.DrainAll(natsAppend))
	assert.ElementsMatch(t, results, nats10Values)
}

func TestStateFlatMap(t *testing.T) {
	sumStream := stream.Sum(nats10)
	ioSum := stream.Head(sumStream)
	sum := UnsafeIO(t, ioSum)
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
	len := UnsafeIO(t, ioLen)
	assert.Equal(t, 100, len)

	filtered := stream.Filter(natsRepeated, isEven)
	sumStream := stream.Sum(filtered)
	ioSum := stream.Head(sumStream)
	sum := UnsafeIO(t, ioSum)
	assert.Equal(t, 550, sum)
}

func TestChunks(t *testing.T) {
	natsBy10 := stream.ChunkN[int](10)(stream.Take(nats, 19))
	nats10to19IO := stream.Head(stream.Drop(natsBy10, 1))
	nats10to19 := UnsafeIO(t, nats10to19IO)
	assert.ElementsMatch(t, []int{11, 12, 13, 14, 15, 16, 17, 18, 19}, nats10to19)
}

func TestGroupBy(t *testing.T) {
	groupedNats := stream.GroupBy(stream.Take(nats, 7), func(i int) int {
		return i / 5
	})
	groupsIO := stream.ToSlice(groupedNats)
	groups := UnsafeIO(t, groupsIO)
	expected := []fun.Pair[int, []int]{
		{V1: 0, V2: []int{1, 2, 3, 4}},
		{V1: 1, V2: []int{5, 6, 7}},
	}
	assert.ElementsMatch(t, expected, groups)
}

func TestGroupByEval(t *testing.T) {
	groupedNats := stream.GroupByEval(stream.Take(nats, 7), func(i int) io.IO[int] {
		return io.Lift(i / 5)
	})
	groupsIO := stream.ToSlice(groupedNats)
	groups := UnsafeIO(t, groupsIO)
	expected := []fun.Pair[int, []int]{
		{V1: 0, V2: []int{1, 2, 3, 4}},
		{V1: 1, V2: []int{5, 6, 7}},
	}
	assert.ElementsMatch(t, expected, groups)
}

func TestGroupByEvalFailed(t *testing.T) {
	groupedNats := stream.GroupByEval(stream.Take(failedStream, 7), func(i int) io.IO[int] {
		return io.Lift(i / 5)
	})
	groupsIO := stream.ToSlice(groupedNats)
	UnsafeIOExpectError(t, errExpected, groupsIO)
}

func TestFailedStream(t *testing.T) {
	failedStream := stream.Eval(io.Fail[int](errExpected))
	ch := make(chan int)
	toChIO := stream.ToChannel(failedStream, ch)
	fromCh := stream.FromChannel(ch)
	sliceIO := stream.ToSlice(fromCh)
	resIO := io.AndThen(toChIO, sliceIO)
	UnsafeIOExpectError(t, errExpected, resIO)
}

func plus(b int, a int) int {
	return a + b
}

func TestFoldLeftEval(t *testing.T) {
	sumIO := stream.FoldLeft(nats10, 0, plus)
	assert.Equal(t, 55, UnsafeIO(t, sumIO))
}

func TestStateFlatMapWithFinishAndFailureHandling(t *testing.T) {
	sumStream := stream.StateFlatMapWithFinishAndFailureHandling(natsAndThenFail, 0,
		func(i, j int) io.IO[fun.Pair[int, stream.Stream[int]]] {
			return io.Lift(fun.NewPair(i+j, stream.Empty[int]()))
		},
		func(s int) stream.Stream[int] {
			return stream.Emit(s)
		},
		func(s int, err error) stream.Stream[int] {
			assert.Equal(t, errExpected, err)
			return stream.Emit(-s)
		},
	)
	sumIO := stream.Head(sumStream)
	sum := UnsafeIO(t, sumIO)
	assert.Equal(t, -55, sum)
}

func TestStateFlatMapWithFinishAndFailureHandling2(t *testing.T) {
	sumStream := stream.StateFlatMapWithFinishAndFailureHandling(nats10, 0,
		func(i, j int) io.IO[fun.Pair[int, stream.Stream[int]]] {
			if i == 10 {
				return io.Fail[fun.Pair[int, stream.Stream[int]]](errExpected)
			} else {
				return io.Lift(fun.NewPair(i+j, stream.Empty[int]()))
			}
		},
		func(s int) stream.Stream[int] {
			return stream.Emit(s)
		},
		func(s int, err error) stream.Stream[int] {
			assert.Equal(t, errExpected, err)
			return stream.Emit(-s)
		},
	)
	sumIO := stream.Head(sumStream)
	sum := UnsafeIO(t, sumIO)
	assert.Equal(t, -45, sum)
}

func TestWrapf(t *testing.T) {
	wrappedNatsAndThenFail := stream.Wrapf(natsAndThenFail, "wrapped")
	wrLastIO := stream.Last(wrappedNatsAndThenFail)
	_, err1 := io.UnsafeRunSync(wrLastIO)
	if assert.Error(t, err1) {
		assert.Contains(t, err1.Error(), "wrapped")
	}
}

func TestSideEval(t *testing.T) {
	sum := 0
	stm := stream.SideEval(nats10, func(i int) io.IOUnit {
		return io.FromPureEffect(func() {
			sum += i
		})
	})
	UnsafeStreamToSlice(t, stm)
	assert.Equal(t, 55, sum)
}
