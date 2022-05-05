package io_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/primetalk/goio/io"

	"github.com/stretchr/testify/assert"
)

func TestPanicRecovery(t *testing.T) {
	a, b, err2 := myFunc()
	assert.Equal(t, 1, a)
	assert.Equal(t, 0, b)
	assert.NotEmpty(t, err2)
	assert.Contains(t, err2.Error(), "CHECK")
}

func myFunc() (a, b int, err error) {
	defer io.RecoverToErrorVar("myFunc", &err)
	a = 1
	if a == 1 {
		err2 := errors.New("TEST ERROR in myFunc. Code = CHECK")
		panic(err2)
	}
	b = 1
	return
}
