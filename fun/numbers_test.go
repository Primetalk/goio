package fun_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	assert.Equal(t, 5, fun.Min(5, 6))
	assert.Equal(t, 6, fun.Max(5, 6))
}
