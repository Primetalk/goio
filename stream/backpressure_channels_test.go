package stream_test

import (
	"errors"
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected error")
var failedStream = stream.Eval(io.Fail[int](errExpected))
var fDrainAllInts = func(stm stream.Stream[int]) io.IOUnit {
	return stream.DrainAll(stm)
}
var fIgnoreHeadInt = func(stm stream.Stream[int]) io.IOUnit {
	return io.Ignore(stream.Head(stm))
}
var fHeadInt = func(stm stream.Stream[int]) io.IO[int] {
	return stream.Head(stm)
}
var fLastInt = func(stm stream.Stream[int]) io.IO[int] {
	return stream.Last(stm)
}
var failStreamIO = func(stm stream.Stream[int]) io.IOUnit {
	return io.Fail[fun.Unit](errExpected)
}

func TestFanOutFiniteSourceNoFailure(t *testing.T) {
	drainAll := stream.FanOut(nats10, fDrainAllInts, fDrainAllInts)
	UnsafeIO(t, drainAll)
}

func TestFanOutFailedStream(t *testing.T) {
	drainAll := stream.FanOut(failedStream, fDrainAllInts, fDrainAllInts)
	UnsafeIOExpectError(t, errExpected, drainAll)
}

func TestFanOutToShortStream(t *testing.T) {
	drainAll := stream.FanOut(nats10, fDrainAllInts, fIgnoreHeadInt)
	UnsafeIO(t, drainAll)
}

func TestFanOutToAllShortStream(t *testing.T) {
	drainAll := stream.FanOut(nats10, fHeadInt, fHeadInt)
	assert.ElementsMatch(t, []int{1, 1}, UnsafeIO(t, drainAll))
}

func TestFanOutToSingleShortStream(t *testing.T) {
	drainAll := stream.FanOut(nats10, fHeadInt, fLastInt)
	assert.ElementsMatch(t, []int{1, 10}, UnsafeIO(t, drainAll))
}

func TestFanOutToFailedStream(t *testing.T) {
	drainAll := stream.FanOut(nats10, failStreamIO, fIgnoreHeadInt)
	UnsafeIOExpectError(t, errExpected, drainAll)
}
