package io

import (
	"errors"

	"github.com/primetalk/goio/fun"
)

// ToChannel saves the value to the channel
func ToChannel[A any](ch chan<- A) func(A) IO[fun.Unit] {
	return func(a A) IO[fun.Unit] {
		return FromUnit(func() (err error) {
			defer fun.RecoverToErrorVar("writing to channel", &err)
			ch <- a
			return nil
		})
	}
}

// MakeUnbufferedChannel allocates a new unbufered channel.
func MakeUnbufferedChannel[A any]() IO[chan A] {
	return Pure(func() chan A {
		return make(chan A)
	})
}

// CloseChannel is an IO that closes the given channel.
func CloseChannel[A any](ch chan<- A) IO[fun.Unit] {
	return FromPureEffect(func() {
		close(ch)
	})
}

// ToChannelAndClose sends the value to the channel and then closes the channel.
func ToChannelAndClose[A any](ch chan<- A) func(A) IO[fun.Unit] {
	return func(a A) IO[fun.Unit] {
		return AndThen(ToChannel(ch)(a), CloseChannel(ch))
	}
}

// FromChannel reads a single value from the channel
func FromChannel[A any](ch chan A) IO[A] {
	return Eval(func() (a A, err error) {
		var ok bool
		a, ok = <-ch
		if !ok {
			err = errors.New("reading from a closed channel")
		}
		return
	})
}
