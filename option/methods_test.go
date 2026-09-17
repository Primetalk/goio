package option_test

import (
	"strconv"
	"testing"

	"github.com/primetalk/goio/option"
	"github.com/stretchr/testify/require"
)

func TestGenericMethods(t *testing.T) {
	result := option.Some(42).
		Map(strconv.Itoa).
		FlatMap(func(value string) option.Option[int] {
			return option.Some(len(value))
		}).
		Match(strconv.Itoa, func() string { return "empty" })

	require.Equal(t, "2", result)
}
