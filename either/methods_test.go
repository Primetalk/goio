package either_test

import (
	"strconv"
	"testing"

	"github.com/primetalk/goio/either"
	"github.com/stretchr/testify/require"
)

func TestGenericFoldMethod(t *testing.T) {
	result := either.Right[error](42).Fold(
		func(err error) string { return err.Error() },
		strconv.Itoa,
	)

	require.Equal(t, "42", result)
}
