package either_test

import (
	"testing"

	"github.com/primetalk/goio/either"
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/option"
	"github.com/stretchr/testify/assert"
)

func TestEither(t *testing.T) {
	assert.Equal(t, "left", either.Fold(either.Left[string, string]("left"), fun.Identity[string], fun.Const[string]("other")))
	assert.Equal(t, "other", either.Fold(either.Right[string]("right"), fun.Identity[string], fun.Const[string]("other")))
	assert.Equal(t, "left", option.Get(either.GetLeft(either.Left[string, string]("left"))))
	assert.Equal(t, "Right", option.Get(either.GetRight(either.Right[string]("Right"))))
	assert.Equal(t, "left", option.IsEmpty(either.GetLeft(either.Right[string]("left"))))
}
