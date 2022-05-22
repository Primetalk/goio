package slice_test

import (
	"testing"

	"github.com/primetalk/goio/slice"
	"github.com/stretchr/testify/assert"
)

func IsEven(i int) bool {
	return i % 2 == 0
}

var nats10Values = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

func TestFilter(t *testing.T) {
	sumEven := slice.Sum(slice.Filter(nats10Values, IsEven))
	assert.Equal(t, 30, sumEven)

	sumOdd := slice.Sum(slice.FilterNot(nats10Values, IsEven))
	assert.Equal(t, 25, sumOdd)
}
