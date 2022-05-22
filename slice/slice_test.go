package slice_test

import (
	"testing"

	"github.com/primetalk/goio/slice"
	"github.com/stretchr/testify/assert"
)

func IsEven(i int) bool {
	return i%2 == 0
}

var nats10Values = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

func TestFilter(t *testing.T) {
	sumEven := slice.Sum(slice.Filter(nats10Values, IsEven))
	assert.Equal(t, 30, sumEven)

	sumOdd := slice.Sum(slice.FilterNot(nats10Values, IsEven))
	assert.Equal(t, 25, sumOdd)
}

func TestFlatten(t *testing.T) {
	floatsNested := slice.Map(nats10Values, func(i int) []float32 {
		return []float32{float32(i), float32(2 * i)}
	})
	floats := slice.Flatten(floatsNested)
	assert.Equal(t, float32(55+55*2), slice.Sum(floats))
}

func TestSet(t *testing.T) {
	intsDuplicated := slice.FlatMap(nats10Values, func(i int) []int {
		return slice.Map(nats10Values, func(j int) int { return i + j })
	})
	intsSet := slice.ToSet(intsDuplicated)
	assert.Equal(t, 19, slice.SetSize(intsSet))
}
