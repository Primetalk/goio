package io

import (
	"errors"
	"fmt"

	"github.com/primetalk/goio/fun"
)

// Continuation represents one step of a multistep computation.
// ObtainResult evaluates explicit continuation chains iteratively. Composition
// combinators may still enter nested ObtainResult calls, so this representation
// is not an unbounded, single-interpreter stack-safety guarantee.
// It is being used to reduce risk of stack overflow.
// It's a universal way to do "trampolining" within goio.
type Continuation[A any] func() ResultOrContinuation[A]

// ResultOrContinuation is either a final result (value or error) or another continuation.
type ResultOrContinuation[A any] struct {
	Value        A
	Error        error
	Continuation *Continuation[A]
}

// MaxContinuationDepth is the default maximum number of continuation functions
// that ObtainResult invokes before giving up. Its initial value is 1,000,000.
// ObtainResult snapshots this value once at the start of each execution.
//
// MaxContinuationDepth remains mutable for compatibility. Callers must
// configure it before concurrent execution starts; reads and external writes
// are not synchronized and concurrent mutation is a data race.
var MaxContinuationDepth = 1_000_000

var nilContinuationIsBeingEnforced = errors.New("nil continuation is being enforced")

// ObtainResult executes continuation functions until a final result is obtained.
// The current MaxContinuationDepth value is captured once per execution. Zero
// and negative limits execute no continuation functions and return a limit
// error. A nil initial or intermediate continuation returns an error.
func ObtainResult[A any](c Continuation[A]) (res A, err error) {
	defer fun.RecoverToErrorVar("ObtainResult", &err)
	if c == nil {
		err = nilContinuationIsBeingEnforced
	} else {
		limit := MaxContinuationDepth
		cont := c
		for i := 0; i < limit; i++ {
			if cont == nil {
				err = nilContinuationIsBeingEnforced
				return
			}
			contResult := cont()
			if contResult.Continuation == nil {
				res = contResult.Value
				err = contResult.Error
				return
			} else {
				cont = *contResult.Continuation
				if cont == nil {
					err = nilContinuationIsBeingEnforced
					return
				}
			}
		}
		err = fmt.Errorf("couldn't enforce continuation in %d iterations", limit)
	}
	return
}
