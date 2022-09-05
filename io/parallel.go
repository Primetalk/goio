package io

import (
	"time"

	"github.com/primetalk/goio/either"
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/slice"
)

// ParallelInExecutionContext starts the given IOs in the provided `ExecutionContext` and waits for all results.
func ParallelInExecutionContext[A any](ec ExecutionContext) func(ios []IO[A]) IO[[]A] {
	return func(ios []IO[A]) IO[[]A] {
		ioFibers := slice.Map(ios, StartInExecutionContext[A](ec))
		fibersIO := Sequence(ioFibers)
		return FlatMap(fibersIO, func(fibers []Fiber[A]) IO[[]A] {
			joins := slice.Map(fibers, func(fiber Fiber[A]) IO[A] { return fiber.Join() })
			return Sequence(joins)
		})
	}
}

// Parallel starts the given IOs in Go routines and waits for all results.
func Parallel[A any](ios ...IO[A]) IO[[]A] {
	return ParallelInExecutionContext[A](globalUnboundedExecutionContext)(ios)
}

// ConcurrentlyFirst - runs all IOs in parallel.
// returns the very first result.
// TODO: after obtaining result - cancel the other IOs.
func ConcurrentlyFirst[A any](ios []IO[A]) IO[A] {
	channelIO := Pure(func() chan GoResult[A] {
		return make(chan GoResult[A], len(ios))
		// we will only read the very first response. Hence the other go routines could hang if sending to unbuffered channel
	})
	return FlatMap(channelIO, func(channel chan GoResult[A]) IO[A] {
		ioSendToChannel := slice.Map(ios, func(ioa IO[A]) IO[fun.Unit] {
			goResult := FoldToGoResult(ioa)
			return FlatMap(goResult, ToChannel(channel))
		})
		parallelSendResults := Parallel(ioSendToChannel...)
		ignoreParallelResultButCloseChannelAfterwards := FireAndForget(AndThen(parallelSendResults, CloseChannel(channel)))
		readFirstFromChannel := FromChannel(channel)
		ignoreParallelResultsAndThenReadFirstFromChannel := AndThen(
			ignoreParallelResultButCloseChannelAfterwards,
			readFirstFromChannel)
		return UnfoldGoResult(ignoreParallelResultsAndThenReadFirstFromChannel)
	})
}

// PairSequentially runs two IOs sequentially and returns both results.
func PairSequentially[A any, B any](ioa IO[A], iob IO[B]) IO[fun.Pair[A, B]] {
	return FlatMap(ioa, func(a A) IO[fun.Pair[A, B]] {
		return Map(iob, func(b B) fun.Pair[A, B] {
			return fun.NewPair(a, b)
		})
	})
}

// PairParallel runs two IOs in parallel and returns both results.
func PairParallel[A any, B any](ioa IO[A], iob IO[B]) IO[fun.Pair[A, B]] {
	return Map(
		Parallel(
			Map(ioa, either.Left[A, B]),
			Map(iob, either.Right[A, B]),
		),
		func(es []either.Either[A, B]) fun.Pair[A, B] {
			if es[0].IsLeft {
				return fun.NewPair(es[0].Left, es[1].Right)
			} else {
				return fun.NewPair(es[1].Left, es[0].Right)
			}
		},
	)
}

// MeasureDuration captures the wall time that was needed to evaluate the given IO.
func MeasureDuration[A any](ioa IO[A]) IO[fun.Pair[A, time.Duration]] {
	return Map(
		PairSequentially(Pure(time.Now), ioa),
		func(p fun.Pair[time.Time, A]) fun.Pair[A, time.Duration] {
			return fun.Pair[A, time.Duration]{
				V1: p.V2,
				V2: time.Since(p.V1),
			}
		},
	)
}

// RunAlso runs the other IO in parallel, but returns only the result of the first IO.
func RunAlso[A any](ioa IO[A], other IOUnit) IO[A] {
	return Map(PairParallel(ioa, other), fun.PairV1[A, fun.Unit])
}
