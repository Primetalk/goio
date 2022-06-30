package fun_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/stretchr/testify/assert"
)

func TestCast(t *testing.T) {
	hello := "hello"
	helloi := fun.CastAsInterface(hello)
	assert.Equal(t, hello, helloi)
	assert.Equal(t, hello, fun.UnsafeCast[string](hello))
	h, err1 := fun.Cast[string](helloi)
	assert.NoError(t, err1)
	assert.Equal(t, hello, h)
}
