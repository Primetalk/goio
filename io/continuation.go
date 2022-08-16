package io

import (
	"errors"
	"fmt"

	"github.com/primetalk/goio/fun"
)

// Continuation represents some multistep computation.
// It is being used to avoid stack overflow. It's a universal way to do "trampolining".
type Continuation[A any] func() ResultOrContinuation[A]

// ResultOrContinuation is either a final result (value or error) or another continuation.
type ResultOrContinuation[A any] struct {
	Value        A
	Error        error
	Continuation *Continuation[A]
}

// MaxContinuationDepth is equal to 1000000000000. It's the maximum depth we run continuation before giving up.
var MaxContinuationDepth = 1000000000000

// ObtainResult executes continuation until final result is obtained.
func ObtainResult[A any](c Continuation[A]) (res A, err error) {
	defer fun.RecoverToErrorVar("ObtainResult", &err)
	if c == nil {
		err = errors.New("nil continuation is being enforced")
	} else {
		cont := c
		for i := 0; i < MaxContinuationDepth; i++ {
			contResult := cont()
			if contResult.Continuation == nil {
				res = contResult.Value
				err = contResult.Error
				return
			} else {
				cont = *contResult.Continuation
			}
		}
		err = fmt.Errorf("couldn't enforce continuation in %d iterations", MaxContinuationDepth)
	}
	return
}
