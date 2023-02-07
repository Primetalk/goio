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

func IsPositive(i int) bool {
	return i > 0
}
func IsNegative(i int) bool {
	return i < 0
}
func add(i, j int) int {
	return i + j
}

var nats10Values = slice.Nats(10)

func TestFilter(t *testing.T) {
	sumEven := slice.Sum(slice.Filter(nats10Values, IsEven))
	assert.Equal(t, 30, sumEven)

	sumOdd := slice.Sum(slice.FilterNot(nats10Values, IsEven))

	assert.Equal(t, 25, sumOdd)
	assert.True(t, slice.Exists(IsEven)(nats10Values))
	assert.False(t, slice.Forall(IsEven)(nats10Values))
	assert.True(t, slice.Forall(IsPositive)(nats10Values))
	assert.False(t, slice.Forall(IsNegative)(nats10Values))

	even, odd := slice.Partition(nats10Values, IsEven)
	assert.Equal(t, sumOdd, slice.Sum(odd))
	assert.Equal(t, sumEven, slice.Sum(even))
}

func TestFlatten(t *testing.T) {
	floatsNested := slice.Map(nats10Values, func(i int) []float32 {
		return []float32{float32(i), float32(2 * i)}
	})
	floats := slice.AppendAll(floatsNested...)
	assert.Equal(t, float32(55+55*2), slice.Sum(floats))
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

func TestCount(t *testing.T) {
	cntEven := slice.Count(nats10Values, IsEven)
	assert.Equal(t, 5, cntEven)
}

func TestReduce(t *testing.T) {
	assert.Equal(t, 55,
		slice.Reduce(slice.Nats(10), add),
	)
}

func TestZipWithIndex(t *testing.T) {
	znats20 := slice.Range(0, 19)
	nats10 := slice.Drop(znats20, 10)
	p1 := slice.ZipWithIndex(nats10)
	p2 := slice.ZipWith(nats10, slice.Take(znats20, 10))
	assert.Equal(t, fun.NewPair(0, 10), p1[0])
	assert.ElementsMatch(t, p1, slice.Map(p2, fun.PairSwap[int, int]))
}

func TestForEach(t *testing.T) {
	s := 0
	slice.ForEach(slice.Nats(10), func(i int) {
		s += i
	})
	assert.Equal(t, 55, s)
}

func TestIndexOf(t *testing.T) {
	assert.Equal(t, 4, slice.IndexOf(nats10Values, 5))
	assert.Equal(t, -1, slice.IndexOf(nats10Values, 11))
}

func TestReverse(t *testing.T) {
	znats0 := slice.Range(0, 0)
	assert.ElementsMatch(t, znats0, slice.Reverse(znats0))
	znats1 := slice.Range(0, 1)
	assert.ElementsMatch(t, znats1, slice.Reverse(znats1))
	znats2 := slice.Range(0, 2)
	assert.ElementsMatch(t, []int{1, 0}, slice.Reverse(znats2))
}

func TestRemove(t *testing.T) {
	znats10 := slice.Range(0, 10)
	znats5 := slice.Range(0, 5)
	znats510 := slice.Range(5, 10)
	assert.ElementsMatch(t, znats510, slice.Remove(znats10, znats5))
}
