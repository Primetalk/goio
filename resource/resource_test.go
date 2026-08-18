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

func TestFailedResource(t *testing.T) {
	expectedErr := errors.New("some error")
	res := resource.Fail[string](expectedErr)
	failed := resource.Use(res, func(s string) io.IO[int] {
		return io.Fail[int](errors.New("some other error"))
	})
	_, err := io.UnsafeRunSync(failed)
	assert.Equal(t, expectedErr, err)
}

func TestChannelResource(t *testing.T) {
	stringChannel := resource.UnbufferedChannel[string]()
	helloIO := resource.Use(stringChannel, func(ch chan string) io.IO[string] {
		notify := io.NotifyToChannel(100*time.Millisecond, "hello", ch)
		return io.AndThen(notify, io.FromChannel(ch))
	})
	hello, err := io.UnsafeRunSync(helloIO)
	assert.NoError(t, err)
	assert.Equal(t, "hello", hello)
}

func TestResourceInResource(t *testing.T) {
	res1 := resource.NewResource(
		io.Lift("resource1"),
		func(s string) io.IO[fun.Unit] {
			assert.Equal(t, "resource1", s)
			return io.IOUnit1
		},
	)
	res2 := resource.FlatMap(res1, func(s string) resource.Resource[fun.Pair[string, string]] {

		return resource.NewResource(
			io.Lift(fun.NewPair(s, "resource2")),
			func(p fun.Pair[string, string]) io.IO[fun.Unit] {
				assert.Equal(t, "resource2", p.V2)
				return io.IOUnit1
			},
		)
	})
	// res2 := resource.FlatMap(resMapped, func(i int) resource.Resource[Pair[int, ]])
	io18 := resource.Use(res2, func(p fun.Pair[string, string]) io.IO[int] {
		return io.Lift(len(p.V1) + len(p.V2))
	})
	res18, err := io.UnsafeRunSync(io18)
	assert.Equal(t, err, nil)
	assert.Equal(t, res18, 18)
}

func TestFlatMapReleasesOuterResourceWhenInnerAcquisitionFails(t *testing.T) {
	innerAcquireErr := errors.New("inner acquisition failed")
	outerReleaseCount := 0
	outer := resource.NewResource(
		io.Lift("outer"),
		func(string) io.IO[fun.Unit] {
			return io.FromPureEffect(func() {
				outerReleaseCount++
			})
		},
	)
	composed := resource.FlatMap(outer, func(string) resource.Resource[int] {
		return resource.Fail[int](innerAcquireErr)
	})
	useCalled := false

	_, err := io.UnsafeRunSync(resource.Use(composed, func(int) io.IO[int] {
		useCalled = true
		return io.Lift(1)
	}))

	assert.Equal(t, innerAcquireErr, err)
	assert.False(t, useCalled)
	assert.Equal(t, 1, outerReleaseCount)
}

func TestFlatMapPreservesInnerAcquisitionErrorWhenOuterReleaseFails(t *testing.T) {
	innerAcquireErr := errors.New("inner acquisition failed")
	outerReleaseErr := errors.New("outer release failed")
	outerReleaseCount := 0
	outer := resource.NewResource(
		io.Lift("outer"),
		func(string) io.IO[fun.Unit] {
			return io.AndThen(
				io.FromPureEffect(func() {
					outerReleaseCount++
				}),
				io.Fail[fun.Unit](outerReleaseErr),
			)
		},
	)
	composed := resource.FlatMap(outer, func(string) resource.Resource[int] {
		return resource.Fail[int](innerAcquireErr)
	})

	_, err := io.UnsafeRunSync(resource.Use(composed, func(value int) io.IO[int] {
		return io.Lift(value)
	}))

	assert.Equal(t, innerAcquireErr, err)
	assert.Equal(t, 1, outerReleaseCount)
}

func TestFlatMapReleasesSuccessfullyAcquiredResourcesInReverseOrder(t *testing.T) {
	releaseOrder := []string{}
	outer := resource.NewResource(
		io.Lift("outer"),
		func(value string) io.IO[fun.Unit] {
			return io.FromPureEffect(func() {
				releaseOrder = append(releaseOrder, value)
			})
		},
	)
	composed := resource.FlatMap(outer, func(string) resource.Resource[string] {
		return resource.NewResource(
			io.Lift("inner"),
			func(value string) io.IO[fun.Unit] {
				return io.FromPureEffect(func() {
					releaseOrder = append(releaseOrder, value)
				})
			},
		)
	})

	value, err := io.UnsafeRunSync(resource.Use(composed, io.Lift[string]))

	assert.NoError(t, err)
	assert.Equal(t, "inner", value)
	assert.Equal(t, []string{"inner", "outer"}, releaseOrder)
}
