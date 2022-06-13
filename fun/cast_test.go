package fun_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/stretchr/testify/assert"
)

func TestCast(t *testing.T) {
	assert.Equal(t, "hello", fun.CastAsInterface("hello"))
	assert.Equal(t, "hello", fun.UnsafeCast[string]("hello"))
}
