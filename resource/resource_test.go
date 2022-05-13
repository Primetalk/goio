package resource_test

import (
	"errors"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/resource"
	"github.com/stretchr/testify/assert"
)

func TestResource(t *testing.T) {
	res := resource.NewResource(
		io.Lift("resource"),
		func(s string) io.IO[io.Unit]{ 
			assert.Equal(t, "resource", s)
			return io.IOUnit1
		},
	)

	io8 := resource.Use(res, func(s string) io.IO[int] {
		return io.Lift(len(s))
	})
	res8, err := io.UnsafeRunSync(io8)
	assert.Equal(t, err, nil)
	assert.Equal(t, res8, 8)
}
func TestResourceFail(t *testing.T) {
	released := false
	res := resource.NewResource(
		io.Lift("resource"),
		func(s string) io.IO[io.Unit]{ 
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
