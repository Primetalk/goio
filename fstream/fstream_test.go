package fstream_test

import (
	"testing"

	"github.com/primetalk/goio/fstream"
	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	empty := fstream.Empty[int]()
	_ = UnsafeIO(t, fstream.DrainAll(empty))
	stream10_12 := fstream.LiftMany(10, 11, 12)
	stream20_24 := Mul2(stream10_12)
	res := UnsafeIO(t, fstream.ToSlice(stream20_24))
	assert.Equal(t, []int{20, 22, 24}, res)
}

func TestGenerate(t *testing.T) {
	powers2 := fstream.Unfold(1, func(s int) int {
		return s * 2
	})

	res := UnsafeIO(t, fstream.Head(powers2))
	assert.Equal(t, 2, res)

	powers2_10 := fstream.Drop(powers2, 9)
	res = UnsafeIO(t, fstream.IOStreamFlatMap(powers2_10, fstream.Head[int]))
	assert.Equal(t, 1024, res)

	res = UnsafeIO(t, fstream.IOStreamFlatMap(fstream.Take(powers2, 10), fstream.Last[int]))
	assert.Equal(t, 1024, res)
}

func TestDrainAll(t *testing.T) {
	results := []int{}
	natsAppend := fstream.MapEval(
		fstream.Take(fstream.Repeat(nats10.ToStream()), 10).ToStream(),
		func(a int) io.IO[int] {
			return io.Eval(func() (int, error) {
				results = append(results, a)
				return a, nil
			})
		})
	_ = UnsafeIO(t, fstream.DrainAll(natsAppend))
	assert.ElementsMatch(t, results, nats10Values)
}

func TestSum(t *testing.T) {
	sumStream := fstream.Sum(nats10.ToStream())
	ioSum := fstream.IOHead(sumStream)
	sum := UnsafeIO(t, ioSum)
	assert.Equal(t, 55, sum)
}

func TestStateFlatMap(t *testing.T) {
	sumStream := fstream.Sum2(nats10.ToStream())
	ioSum := fstream.IOHead(sumStream)
	sum := UnsafeIO(t, ioSum)
	assert.Equal(t, 55, sum)
}

func isEven(i int) bool {
	return i%2 == 0
}

func plus(b int, a int) int {
	return a + b
}
