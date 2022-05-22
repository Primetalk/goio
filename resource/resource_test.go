package resource_test

import (
	"errors"
	"testing"
	"time"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/resource"
	"github.com/stretchr/testify/assert"
)

func TestResource(t *testing.T) {
	res := resource.NewResource(
		io.Lift("resource"),
		func(s string) io.IO[fun.Unit] {
			assert.Equal(t, "resource", s)
			return io.IOUnit1
		},
	)
	resMapped := resource.Map(res, func(s string) int {
		return len(s)
	})
	// res2 := resource.FlatMap(resMapped, func(i int) resource.Resource[Pair[int, ]])
	io9 := resource.Use(resMapped, func(i int) io.IO[int] {
		return io.Lift(i + 1)
	})
	res9, err := io.UnsafeRunSync(io9)
	assert.Equal(t, err, nil)
	assert.Equal(t, res9, 9)
}
func TestResourceFail(t *testing.T) {
	released := false
	res := resource.NewResource(
		io.Lift("resource"),
		func(s string) io.IO[fun.Unit] {
			assert.Equal(t, "resource", s)
			released = true
			return io.IOUnit1
		},
	)

	failed := resource.Use(res, func(s string) io.IO[int] {
		return io.Fail[int](errors.New("some error"))
	})
	_, err := io.UnsafeRunSync(failed)
	assert.NotEqual(t, err, nil)
	assert.True(t, released)
}

func TestChannelResource(t *testing.T) {
	stringChannel := resource.UnbufferedChannel[string]()
	helloIO := resource.Use(stringChannel, func(ch chan string) io.IO[string] {
		notify := io.NotifyToChannel(100 * time.Millisecond, "hello", ch)
		return io.AndThen(notify, io.FromChannel(ch))
	})
	hello, err := io.UnsafeRunSync(helloIO)
	assert.NoError(t, err)
	assert.Equal(t, "hello", hello)
}
