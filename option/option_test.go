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
var onone = option.None[string]()

func TestMap(t *testing.T) {
	assert.Equal(t, option.Some(5), option.Map(ohello, StringLen))
	assert.Equal(t, option.None[int](), option.Map(onone, StringLen))
}

func Contains(substring string) func(s string) bool {
	return func(s string) bool {
		return strings.Contains(s, substring)
	}
}

func TestFilter(t *testing.T) {
	assert.Equal(t, ohello, option.Filter(ohello, Contains("llo")))
	assert.Equal(t, onone, option.Filter(onone, Contains("llo")))
}

func TestIsDefined(t *testing.T) {
	assert.True(t, option.IsDefined(ohello))
	assert.True(t, option.IsEmpty(onone))
}

func TestFlatten(t *testing.T) {
	assert.Equal(t, "hello", option.Get(option.Flatten(option.Some(ohello))))
	assert.Panics(t, func() { option.Get(option.Flatten(option.Some(onone))) })
}
func TestForEach(t *testing.T) {
	option.ForEach(ohello, func(s string) {
		assert.Equal(t, "hello", s)
	})
	option.ForEach(onone, func(s string) {
		assert.Fail(t, "unexpected call")
	})
}
