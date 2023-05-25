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

func TestIntersection(t *testing.T) {
	znats10 := slice.Range(0, 10)
	znats515 := slice.Range(5, 15)
	znats5 := slice.Range(5, 10)
	assert.ElementsMatch(t, znats5, slice.Intersection(znats10, znats515))
}

func TestUnion(t *testing.T) {
	znats10 := slice.Range(0, 10)
	znats515 := slice.Range(5, 15)
	znats15 := slice.Range(0, 15)
	assert.ElementsMatch(t, znats15, slice.Union(znats10, znats515))
}

func TestHeadTail(t *testing.T) {
	znats5 := slice.Range(0, 5)
	z, nats5 := slice.HeadTail(znats5)
	assert.Equal(t, 0, z)
	assert.ElementsMatch(t, slice.Range(1, 5), nats5)
}

func TestHeadTailPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != slice.ErrHeadOfEmptySlice {
			t.Errorf("expected panic ErrHeadOfEmptySlice")
		}
	}()
	znats5 := slice.Range(0, 0)
	slice.HeadTail(znats5)
}

func TestHeadAndTail(t *testing.T) {
	znats5 := slice.Range(0, 5)
	z := slice.Head(znats5)
	nats5 := slice.Tail(znats5)
	assert.Equal(t, 0, z)
	assert.ElementsMatch(t, slice.Range(1, 5), nats5)
}

func TestDistinct(t *testing.T) {
	ints := []int{2, 2, 0, 1, 1, 5, 0, 2}
	assert.ElementsMatch(t, slice.Distinct(ints), []int{2, 0, 1, 5})
}

func TestIntersperse(t *testing.T) {
	ints := []int{2, 2, 0, 1}
	assert.ElementsMatch(t, slice.Intersperse(ints, -1), []int{2, -1, 2, -1, 0, -1, 1})
	ints2 := []int{2}
	assert.ElementsMatch(t, slice.Intersperse(ints2, -1), []int{2})
	ints3 := []int{}
	assert.ElementsMatch(t, slice.Intersperse(ints3, -1), []int{})
}

func TestBuildIndex(t *testing.T) {
	strings := []string{"a", "four", "eleven", "five"}
	l := func(s string) int { return len(s) }
	index := slice.BuildIndex(strings, l)
	assert.ElementsMatch(t, []string{"a"}, index[1])
	assert.ElementsMatch(t, []string{"four", "five"}, index[4])
}

func BuildUniqueIndex(t *testing.T) {
	strings := []string{"a", "four", "eleven", "five"}
	l := func(s string) int { return len(s) }
	index := slice.BuildUniqueIndex(strings, l)
	assert.Equal(t, "a", index[1])
	assert.Equal(t, "five", index[4])
}
