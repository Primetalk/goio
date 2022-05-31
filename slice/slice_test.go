package slice_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
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
	floats := slice.AppendAll(floatsNested...)
	assert.Equal(t, float32(55+55*2), slice.Sum(floats))
}

func TestSet(t *testing.T) {
	intsDuplicated := slice.FlatMap(nats10Values, func(i int) []int {
		return slice.Map(nats10Values, func(j int) int { return i + j })
	})
	intsSet := slice.ToSet(intsDuplicated)
	assert.Equal(t, 19, slice.SetSize(intsSet))
}

func TestGroupBy(t *testing.T) {
	intsDuplicated := slice.FlatMap(nats10Values, func(i int) []int {
		return slice.Map(nats10Values, func(j int) int { return i + j })
	})
	intsGroups := slice.GroupBy(intsDuplicated, fun.Identity[int])
	assert.Equal(t, 19, len(intsGroups))
	for k, as := range intsGroups {
		assert.Equal(t, k, as[0])
	}
}

func TestSliding(t *testing.T) {
	intWindows := slice.Sliding(nats10Values, 3, 2)
	assert.ElementsMatch(t, intWindows[4], []int{9, 10})
	intWindows = slice.Sliding(nats10Values, 2, 5)
	assert.ElementsMatch(t, intWindows[1], []int{6, 7})
}

func TestGrouped(t *testing.T) {
	intWindows := slice.Grouped(nats10Values, 3)
	assert.ElementsMatch(t, intWindows[3], []int{10})
}

func TestGroupByMapCount(t *testing.T) {
	counted := slice.GroupByMapCount(nats10Values, IsEven)
	assert.Equal(t, 5, counted[false])
	assert.Equal(t, 5, counted[true])
}
