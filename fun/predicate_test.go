package fun_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/stretchr/testify/assert"
)

func TestEqualTo(t *testing.T) {
	assert.True(t, fun.IsEqualTo(5)(5))
	assert.False(t, fun.Not(fun.IsEqualTo(5))(5))
}
