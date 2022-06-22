package option_test

import (
	"strings"
	"testing"

	"github.com/primetalk/goio/option"
	"github.com/stretchr/testify/assert"
)

func StringLen(s string) int {
	return len(s)
}

var ohello = option.Some("hello")

func TestMap(t *testing.T) {
	assert.Equal(t, option.Some(5), option.Map(ohello, StringLen))
}

func Contains(substring string) func(s string) bool {
	return func(s string) bool {
		return strings.Contains(s, substring)
	}
}

func TestFilter(t *testing.T) {
	assert.Equal(t, ohello, option.Filter(ohello, Contains("llo")))
}

func TestFlatten(t *testing.T) {
	assert.Equal(t, "hello", option.Get(option.Flatten(option.Some(ohello))))

}
