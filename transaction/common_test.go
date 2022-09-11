package transaction_test

import (
	"errors"
	"testing"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected error")

var failure = io.Fail[string](errExpected)

func UnsafeIO[A any](t *testing.T, ioa io.IO[A]) A {
	res, err1 := io.UnsafeRunSync(ioa)
	assert.NoError(t, err1)
	return res
}

func UnsafeIOExpectError[A any](t *testing.T, expected error, ioa io.IO[A]) {
	_, err1 := io.UnsafeRunSync(ioa)
	if assert.Error(t, err1) {
		assert.Equal(t, expected, err1)
	}
}
