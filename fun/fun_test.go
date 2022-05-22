package fun_test

import (
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/stretchr/testify/assert"
)


func TestFun(t *testing.T) {
	assert.Equal(t, "hello", fun.ConstUnit("hello")(fun.Unit1))
	assert.Equal(t, "hello", fun.Identity("hello"))
	concat := func(a string)func(string)string {
		return func (b string) string{
			return a + b
		}
	}
	assert.Equal(t, "ab", concat("a")("b"))
	assert.Equal(t, "ba", fun.Swap(concat)("a")("b"))
}

func TestPair(t *testing.T){
	assert.Equal(t, "a", fun.NewPair("a", "b").V1)
	assert.Equal(t, "b", fun.NewPair("a", "b").V2)
}

func TestEither(t *testing.T){
	assert.Equal(t, "left", fun.Fold(fun.Left[string, string]("left"), fun.Identity[string], fun.Const[string]("other")))
	assert.Equal(t, "other", fun.Fold(fun.Right[string]("right"), fun.Identity[string], fun.Const[string]("other")))
}

func TestToString(t *testing.T){
	assert.Equal(t, "1", fun.ToString(1))
}
