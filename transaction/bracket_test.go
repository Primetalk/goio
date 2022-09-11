package transaction_test

import (
	"fmt"
	"testing"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/transaction"
	"github.com/stretchr/testify/assert"
)

func TestBracket(t *testing.T) {
	createVarIO := io.Pure(func() *int {
		i := 0
		return &i
	})
	pitoa := func(pint *int) io.IO[string] {
		return io.Pure(func() string {
			return fmt.Sprintf("%d", *pint)
		})
	}
	bracketedPitoaIO := transaction.Bracket[string](createVarIO, fun.Const[*int](io.IOUnit1), fun.Const[*int](io.IOUnit1))(pitoa)
	assert.Equal(t, "0", UnsafeIO(t, bracketedPitoaIO))
}

func TestBracketRollback(t *testing.T) {
	createVarIO := io.Pure(func() *int {
		i := 0
		return &i
	})

	bracketedPitoaIO := transaction.Bracket[string](createVarIO, fun.Const[*int](io.IOUnit1), fun.Const[*int](io.IOUnit1))(fun.Const[*int](failure))
	UnsafeIOExpectError(t, errExpected, bracketedPitoaIO)
}
