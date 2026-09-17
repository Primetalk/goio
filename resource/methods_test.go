package resource_test

import (
	"strconv"
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/resource"
	"github.com/stretchr/testify/require"
)

func TestGenericMethods(t *testing.T) {
	released := false
	res := resource.NewResource(
		io.Lift(42),
		func(int) io.IO[fun.Unit] {
			return io.FromPureEffect(func() { released = true })
		},
	).
		Map(strconv.Itoa).
		FlatMap(func(value string) resource.Resource[int] {
			return resource.NewResource(
				io.Lift(len(value)),
				func(int) io.IO[fun.Unit] { return io.IOUnit1 },
			)
		})

	result, err := io.UnsafeRunSync(res.Use(func(value int) io.IO[string] {
		return io.Lift(strconv.Itoa(value))
	}))
	require.NoError(t, err)
	require.Equal(t, "2", result)
	require.True(t, released)
}
