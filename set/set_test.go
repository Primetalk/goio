package set_test

import (
	"testing"

	"github.com/primetalk/goio/set"
	"github.com/primetalk/goio/slice"
	"github.com/stretchr/testify/assert"
)

var nats10Values = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

func TestSet(t *testing.T) {
	intsDuplicated := slice.FlatMap(nats10Values, func(i int) []int {
		return slice.Map(nats10Values, func(j int) int { return i + j })
	})
	intsSet := slice.ToSet(intsDuplicated)
	assert.Equal(t, 19, set.SetSize(intsSet))
}

func TestFilter(t *testing.T) {
	s123 := slice.ToSet([]int{1, 2, 3})
	assert.True(t, set.Contains(s123)(3), "3 \\in {1,2,3}")
	sl13 := slice.Filter([]int{1, 3, 5, 7}, set.Contains(s123))
	assert.ElementsMatch(t, []int{1, 3}, sl13)
}
