package io

import "github.com/primetalk/goio/fun"

// ToChannel saves the value to the channel
func ToChannel[A any](ch chan A) func(A) IO[fun.Unit] {
	return func(a A) IO[fun.Unit] {
		return FromUnit(func() error {
			ch <- a
			return nil
		})
	}
}

// CloseChannel is an IO that closes the given channel.
func CloseChannel[A any](ch chan A) IO[fun.Unit] {
	return FromUnit(func() error {
		close(ch)
		return nil
	})
}
// ToChannelAndClose sends the value to the channel and then closes the channel.
func ToChannelAndClose[A any](ch chan A) func(A) IO[fun.Unit] {
	return func(a A) IO[fun.Unit] {
		return AndThen(ToChannel(ch)(a), CloseChannel(ch))
	}
}

// FromChannel reads a single value from the channel
func FromChannel[A any](ch chan A) IO[A] {
	return Pure(func() A {
		return <-ch
	})
}
