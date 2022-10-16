package fun_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/stretchr/testify/assert"
)

func TestPairBoth(t *testing.T) {
	one, a := fun.PairBoth(fun.NewPair(1, "A"))
	assert.Equal(t, 1, one)
	assert.Equal(t, "A", a)
}

func TestPairSwap(t *testing.T) {
	sw := fun.PairSwap(fun.NewPair(1, "A"))
	assert.Equal(t, fun.NewPair("A", 1), sw)
}
