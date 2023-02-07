package fun_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/stretchr/testify/assert"
)

func TestMemoize(t *testing.T) {
	counter := 0
	inc := func(i int) int {
		counter += 1
		return i + 1
	}
	incm := fun.Memoize(inc)
	assert.Equal(t, 2, incm(1))
	assert.Equal(t, 2, incm(1))
	assert.Equal(t, 3, incm(2))
	assert.Equal(t, 2, counter)
}
