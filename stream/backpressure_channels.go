package stream

import (
	"fmt"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/slice"
)

// BackpressureChannel has a control mechanism that allows consumer to
// influence the producer.
// There is a back pressure channel.
// Protocol:
//  sender              |  receiver
//  --------------------+------------------------------------
//                      |   send "Ready to receive" to back channel
//   read back          |   immediately start listening data.
//   if ready           |
//   send data          |   read data
//                      |    start processing
//                      |     the result of processing (ready-to-receive/finished/error)
//   loop               |   LOOP.
//                      |
//                      |    on error after processing
//                      |     send error to back
//                      |    on processing complete
//                      |     send finished to back
//  when finishing:     |
//   send finish signal | on receiving finish signal, stop the loop.
// and read back        |
//                      |
//  when error:         |
//   send error         | on receiving error, stop the loop.
// and read back        |
//                      |
//   if not ready,      |
//   don't send data    |
//   on back error - fail all
//   on back finish - unsubscribe
type BackpressureChannel[A any] struct {
	data chan StreamEvent[A]
	back chan StreamEvent[fun.Unit]
}

var errDataChannelIsClosed = fmt.Errorf("data channel is closed unexpectedly")
var errBackChannelIsClosedTooEarly = fmt.Errorf("couldn't read from BackpressureChannel.back channel on main stream completion")
var errBackChannelNoTerminationConfirmation = fmt.Errorf("protocol error: haven't received termination confirmation")
var errFinishedError = fmt.Errorf("not-an-error: all receivers have unsubscribed")

func NewBackpressureChannel[A any]() BackpressureChannel[A] {
	return BackpressureChannel[A]{
		data: make(chan StreamEvent[A]),
		back: make(chan StreamEvent[fun.Unit]),
	}
}

func (bc BackpressureChannel[A]) SendValue(a A) (bool, error) {
	return bc.Send(StreamEvent[A]{Value: a})
}

func (bc BackpressureChannel[A]) SendError(err error) (bool, error) {
	return bc.Send(StreamEvent[A]{Error: err})
}

// Send receives readiness signal from `back`.
// If ready, sends data to `data`.
func (bc BackpressureChannel[A]) Send(sea StreamEvent[A]) (isFinished bool, err error) {
	u := <-bc.back
	isFinished = u.IsFinished
	err = u.Error
	if err == nil && !u.IsFinished {
		bc.data <- sea
	}
	return
}

func (bc BackpressureChannel[A]) Close() (err error) {
	defer fun.RecoverToErrorVar("close BackpressureChannel", &err)
	close(bc.data)
	last, ok := <-bc.back
	if ok {
		err = last.Error
	} else {
		err = errBackChannelIsClosedTooEarly
	}
	if err == nil {
		if !last.IsFinished {
			err = errBackChannelNoTerminationConfirmation
		}
	}
	return
}

func (bc BackpressureChannel[A]) CloseReceiverWithError(err error) {
	bc.back <- NewStreamEventError[fun.Unit](err)
	close(bc.back)
}

func (bc BackpressureChannel[A]) CloseReceiverNormally() {
	bc.back <- NewStreamEventFinished[fun.Unit]()
	close(bc.back)
}

// RequestOneItem - sends notification to backpressure channel and receives one item from data channel.
func (bc BackpressureChannel[A]) RequestOneItem() StreamEvent[A] {
	bc.back <- NewStreamEvent(fun.Unit1)
	d, ok := <-bc.data
	if !ok {
		d.Error = errDataChannelIsClosed
	}
	return d
}

// HappyPathReceive forms a stream of a happy path.
func (bc BackpressureChannel[A]) HappyPathReceive() Stream[A] {
	return FromStepResult(
		io.Eval(func() (sra StepResult[A], err error) {
			d := bc.RequestOneItem()
			if err == nil {
				if d.IsFinished {
					sra = NewStepResultFinished[A]()
				} else {
					sra = NewStepResult(d.Value, bc.HappyPathReceive())
				}
			}
			return
		}),
	)
}

// ToBackPressureChannels sends each element to all channels.
func ToBackPressureChannels[A any](stm Stream[A], channels ...BackpressureChannel[A]) io.IO[fun.Unit] {
	streamEvents := ToStreamEvent(stm) // This stream should never fail at the level of io.
	empty := Empty[[]BackpressureChannel[A]]()
	stmUnits := StateFlatMapWithFinish(
		streamEvents,
		channels,
		func(sea StreamEvent[A], channels []BackpressureChannel[A]) io.IO[fun.Pair[[]BackpressureChannel[A], Stream[[]BackpressureChannel[A]]]] {
			return io.Eval(func() (p fun.Pair[[]BackpressureChannel[A], Stream[[]BackpressureChannel[A]]], err error) {
				channels2 := make([]BackpressureChannel[A], 0, len(channels))
				for _, ch := range channels {
					var isFinished bool
					isFinished, err = ch.Send(sea)
					if err == nil {
						if !isFinished {
							channels2 = append(channels2, ch)
						}
					} else {
						break
					}
				}
				p = fun.NewPair(channels2, empty)
				if sea.Error == nil {
					if sea.IsFinished {
						// do nothing, this was the last element
					} else {
						if err == nil {
							if len(channels2) > 0 {
								// continue processing.
							} else {
								// should stop processing
								err = errFinishedError
							}
						}
					}
				} else {
					err = sea.Error
				}
				return
			})
		},
		func(channels []BackpressureChannel[A]) Stream[[]BackpressureChannel[A]] {
			return Lift(channels)
		},
	)

	sendAll := io.Recover(Head(stmUnits), func(err error) (res io.IO[[]BackpressureChannel[A]]) {
		if err == errFinishedError {
			res = io.Lift([]BackpressureChannel[A]{})
		} else {
			res = io.Fail[[]BackpressureChannel[A]](err)
		}
		return
	})
	// return sendAll
	sendAndCloseChannels := io.Fold(
		sendAll,
		func(channels2 []BackpressureChannel[A]) io.IOUnit {
			return io.Ignore(io.Parallel(
				slice.Map(channels2, func(bc BackpressureChannel[A]) io.IOUnit {
					return io.FromUnit(func() error {
						return bc.Close()
					})
				})...,
			))
		},
		func(err error) io.IOUnit {
			return io.Fail[fun.Unit](err)
		},
	)

	return sendAndCloseChannels
	//io.Finally(
	// 	sendAll,
	// 	io.Ignore(closeChannels),
	// )
}

// FromBackpressureChannel forms a stream[A] that will be consumed by `f`.
// The result of `f` will be used to report back failures and finish signals.
// this is intended to be run in
func FromBackpressureChannel[A any, B any](bc BackpressureChannel[A], f func(Stream[A]) io.IO[B]) io.IO[B] {
	return io.Fold(
		f(bc.HappyPathReceive()),
		func(b B) io.IO[B] {
			return io.Pure(func() B {
				bc.CloseReceiverNormally()
				return b
			})
		},
		func(err error) io.IO[B] {
			return io.Eval(func() (b1 B, err1 error) {
				bc.CloseReceiverWithError(err)
				err1 = err
				return
			})
		},
	)
}
