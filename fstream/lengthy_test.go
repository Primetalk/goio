package fstream_test

import (
	"testing"

	"github.com/primetalk/goio/fstream"
	"github.com/stretchr/testify/assert"
)

func TestTakeWhile(t *testing.T) {
	nats1112 := fstream.TakeWhile(
		fstream.DropWhile(
			nats,
			func(i int) bool { return i < 10 },
		).ToStream(),
		func(i int) bool { return i < 12 },
	)
	res := UnsafeIOStreamToSlice(t, nats1112)
	assert.ElementsMatch(t, res, []int{10, 11})
}
