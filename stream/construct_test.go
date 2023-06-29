package stream_test

import (
	"testing"

	"github.com/primetalk/goio/stream"
	"github.com/stretchr/testify/assert"
)

var nats = stream.Nats()
var nats10 = stream.Take(nats, 10)
var nats10Values = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

var fibs01 = stream.Fib(0, 1)

func TestNats(t *testing.T) {
	assert.ElementsMatch(t, nats10Values, UnsafeStreamToSlice(t, nats10))
}

func pow2(i int64) int {
	return int(i * i)
}
func TestFibs(t *testing.T) {
	var fibs5 = stream.Take(fibs01, 5)
	var fibs5Values = []int64{1, 1, 2, 3, 5}
	assert.ElementsMatch(t, fibs5Values, UnsafeStreamToSlice(t, fibs5))

	powered := stream.Map(fibs01, pow2)
	filtered := stream.FilterNot(powered, isEven)
	filtered5 := stream.Take(filtered, 5)
	expected := []int{1, 1, 9, 25, 169}
	assert.ElementsMatch(t, expected, UnsafeStreamToSlice(t, filtered5))
	hIO := stream.Head(stream.Drop(fibs01, 55))
	assert.Equal(t, int64(225851433717), UnsafeIO(t, hIO))
}
