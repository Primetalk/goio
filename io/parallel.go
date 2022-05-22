package io

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/slice"
)

// Parallel starts the given IOs in Go routines and waits for all results
func Parallel[A any](ios []IO[A]) IO[[]A] {
	ioFibers := slice.Map(ios, Start[A])
	fibersIO := Sequence(ioFibers)
	return FlatMap(fibersIO, func(fibers []Fiber[A]) IO[[]A] {
		joins := slice.Map(fibers, func(fiber Fiber[A]) IO[A] { return fiber.Join() })
		return Sequence(joins)
	})
}

// ConcurrentlyFirst - runs all IOs in parallel.
// returns the very first result.
// TODO: after obtaining result - cancel the other IOs.
func ConcurrentlyFirst[A any](ios []IO[A]) IO[A] {
	channel := make(chan GoResult[A])
	ioSendToChannelAndCloseChannels := slice.Map(ios, func(ioa IO[A]) IO[fun.Unit] {
		goResult := FoldToGoResult(ioa)
		return FlatMap(goResult, ToChannelAndClose(channel))
	})
	parallelSendResults := Parallel(ioSendToChannelAndCloseChannels)
	ignoreParallelResults := FireAndForget(parallelSendResults)
	readFromChannel := FromChannel(channel)
	ignoreParallelResultsAndThenReadFromChannel := AndThen(ignoreParallelResults, readFromChannel)
	return UnfoldGoResult(ignoreParallelResultsAndThenReadFromChannel)
}
