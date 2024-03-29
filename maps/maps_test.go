package maps_test

import (
	"testing"

	"github.com/primetalk/goio/maps"
	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	m := map[int]int{1: 1, 2: 1, 3: 4}
	assert.ElementsMatch(t, maps.Keys(m), []int{1, 2, 3})
}

func TestMerge(t *testing.T) {
	m1 := map[int]int{1: 1, 2: 1, 3: 4}
	m2 := map[int]int{2: 2, 3: 5, 4: 6}
	m := maps.Merge(m1, m2, func(i, j int) int {
		return i + j
	})
	assert.Equal(t, map[int]int{1: 1, 2: 3, 3: 9, 4: 6}, m)
}

func isEven(i int) bool {
	return i%2 == 0
}

func add(a, b int) int {
	return a + b
}

func TestMapKeys(t *testing.T) {
	m1 := map[int]int{1: 1, 2: 1, 3: 4}
	m2 := maps.MapKeys(m1, isEven, add)
	assert.Equal(t, map[bool]int{true: 1, false: 5}, m2)
}
