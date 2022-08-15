package either_test

import (
	"testing"

	"github.com/primetalk/goio/either"
	"github.com/primetalk/goio/fun"
	"github.com/stretchr/testify/assert"
)

func TestEither(t *testing.T) {
	assert.Equal(t, "left", either.Fold(either.Left[string, string]("left"), fun.Identity[string], fun.Const[string]("other")))
	assert.Equal(t, "other", either.Fold(either.Right[string]("right"), fun.Identity[string], fun.Const[string]("other")))
}
