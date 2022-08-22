package stream

import "github.com/primetalk/goio/io"

// Fields should be checked in order - If Error == nil, If !IsFinished, then Value
type StreamEvent[A any] struct {
	Error      error
	IsFinished bool // true when stream has completed
	Value      A
}

// ToStreamEvent converts the given stream to a stream of StreamEvents.
// Each normal element will become a StreamEvent with data.
// On a failure or finish a single element is returned before the end of the stream.
func ToStreamEvent[A any](stm Stream[A]) Stream[StreamEvent[A]] {
	return Stream[StreamEvent[A]](
		io.Fold(
			io.IO[StepResult[A]](stm),
			func(sra StepResult[A]) io.IO[StepResult[StreamEvent[A]]] {
				var res StepResult[StreamEvent[A]]
				if sra.IsFinished {
					res = NewStepResult(StreamEvent[A]{IsFinished: true}, Empty[StreamEvent[A]]())
				} else if sra.HasValue {
					res = NewStepResult(StreamEvent[A]{Value: sra.Value}, ToStreamEvent(sra.Continuation))
				} else {
					res = NewStepResultEmpty(ToStreamEvent(sra.Continuation))
				}
				return io.Lift(res)
			},
			func(err error) io.IO[StepResult[StreamEvent[A]]] {
				res := NewStepResult(StreamEvent[A]{Error: err}, Empty[StreamEvent[A]]())
				return io.Lift(res)
			},
		),
	)
}
