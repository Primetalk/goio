package contextio

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidEffect is returned when the zero Effect value is executed.
	ErrInvalidEffect = errors.New("contextio: invalid effect")
	// ErrNilContext is returned when Run is called with a nil context.
	ErrNilContext = errors.New("contextio: nil context")
	// ErrEmptyRace is returned when Race has no competitors.
	ErrEmptyRace = errors.New("contextio: empty race")
)

// PanicError reports a panic recovered at an effect boundary.
// Value is the value supplied to panic.
type PanicError struct {
	Value any
}

// Error implements error without invoking formatting methods on the recovered
// value, which could themselves panic.
func (e *PanicError) Error() string {
	return fmt.Sprintf("contextio: recovered panic of type %T", e.Value)
}

// CombinedError preserves a primary computation failure together with a
// secondary finalizer failure.
type CombinedError struct {
	Primary   error
	Secondary error
}

// Error implements error.
func (e *CombinedError) Error() string {
	return fmt.Sprintf("contextio: primary error: %v; release error: %v", e.Primary, e.Secondary)
}

// Unwrap exposes the primary error for errors.Is and errors.As.
func (e *CombinedError) Unwrap() error {
	return e.Primary
}
