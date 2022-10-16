package slice_test

import (
	"testing"

	"github.com/primetalk/goio/slice"
	"github.com/stretchr/testify/assert"
)

func TestRange(t *testing.T) {
	assert.ElementsMatch(t, []int{1, 2, 3, 4}, slice.Nats(4))
}
