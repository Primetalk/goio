# Implementation of IO, Stream, Fiber using go1.18 generics
![Coverage](https://img.shields.io/badge/Coverage-86.9%25-brightgreen)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/56db71f0cf6d4c76b796af26a1d7ef41)](https://app.codacy.com/gh/Primetalk/goio?utm_source=github.com&utm_medium=referral&utm_content=Primetalk/goio&utm_campaign=Badge_Grade_Settings)
[![Go Reference](https://pkg.go.dev/badge/github.com/primetalk/goio.svg)](https://pkg.go.dev/github.com/primetalk/goio)
[![GoDoc](https://godoc.org/github.com/primetalk/goio?status.svg)](https://godoc.org/github.com/primetalk/goio)
[![Go Report Card](https://goreportcard.com/badge/github.com/primetalk/goio)](https://goreportcard.com/report/github.com/primetalk/goio)
[![Version Badge](https://img.shields.io/github/v/tag/primetalk/goio)](https://img.shields.io/github/v/tag/primetalk/goio)
![Go](https://github.com/primetalk/goio/workflows/Go/badge.svg?branch=master)
[![codecov](https://codecov.io/gh/Primetalk/goio/branch/master/graph/badge.svg?token=WXVKKB4EWO)](https://codecov.io/gh/Primetalk/goio)

This library is an attempt to fill the gap of a decent generics streaming libraries in Go lang. The existing alternatives do not yet use Go 1.18 generics to their full potential.

The design is inspired by awesome Scala libraries [cats-effect](https://typelevel.org/cats-effect/) and [fs2](https://fs2.io/).

## Functions

This package provides a few general functions that are sometimes useful.

- `fun.Const[A any, B any](b B)func(A)B`
- `fun.ConstUnit[B any](b B) func(Unit)B`
- `fun.Identity[A any](a A) A` - Identity function returns the given value unchanged.
- `fun.Swap[A any, B any, C any](f func(a A)func(b B)C) func(b B)func(a A)C`
- `fun.Curry[A any, B any, C any](f func(a A, b B)C) func(a A)func(b B)C`
- `fun.Compose[A any, B any, C any](f func(A) B, g func (B) C) func (A) C` - Compose executes the given functions in sequence.

There are also basic data structures - Unit, Pair and Either.

- `fun.Unit` - type that has only one instance
- `fun.Unit1` - the instance of the Unit type
- `fun.Pair[A any, B any]` - type that represents both A and B.
- `fun.Either[A any, B any]` - type that represents either A or B.

For `Either` there are a few helper functions:

- `fun.Left[A any, B any](a A) Either[A, B]`
- `fun.Right[A any, B any](b B) Either[A, B]`
- `fun.IsLeft[A any, B any](eab Either[A, B]) bool`
- `fun.IsRight[A any, B any](eab Either[A, B]) bool`
- `fun.Fold[A any, B any, C any](eab Either[A, B], left func(A)C, right func(B)C) C` - Fold pattern matches Either with two given pattern match handlers

For debug purposes it's useful to convert arbitrary data to strings.

- `fun.ToString[A any](a A) string` - converts the value to string using `Sprintf` `%v`.

## IO

IO encapsulates a calculation and provides a mechanism to compose a few calculations (flat map or bind).

### Construction

To construct an IO one may use the following functions:

- `io.Lift[A any](a A) IO[A]` - lifts a plain value to IO
- `io.Fail[A any](err error) IO[A]` - lifts an error to IO
- `io.FromConstantGoResult[A any](gr GoResult[A]) IO[A]` - FromConstantGoResult converts an existing GoResult value into an IO. Important! This is not for normal delayed IO execution. It cannot provide any guarantee for the moment when this go result was evaluated in the first place. This is just a combination of Lift and Fail.
- `io.Eval[A any](func () (A, error)) IO[A]` - lifts an arbitrary computation. Panics are handled and represented as errors.
- `io.FromPureEffect(f func())IO[fun.Unit]` - FromPureEffect constructs IO from the simplest function signature.
- `io.Delay[A any](f func()IO[A]) IO[A]` - represents a function as a plain IO
- `io.Fold[A any, B any](io IO[A], f func(a A)IO[B], recover func (error)IO[B]) IO[B]` - handles both happy and sad paths.
- `io.Recover[A any](io IO[A], recover func(err error)IO[A])IO[A]` - handle only sad path and recover some errors to happy path.

### Manipulation

The following functions could be used to manipulate computations:

- `io.FlatMap[A any, B any](ioa IO[A], f func(A)IO[B]) IO[B]`
- `io.AndThen[A any, B any](ioa IO[A], iob IO[B]) IO[B]` - AndThen runs the first IO, ignores it's result and then runs the second one.
- `io.Map[A any, B any](ioA IO[A], f func(a A) B) IO[B]`
- `io.MapErr[A any, B any](ioA IO[A], f func(a A) (B, error)) IO[B]`
- `io.Sequence[A any](ioas []IO[A]) (res IO[[]A])`
- `io.SequenceUnit(ious []IO[Unit]) (res IOUnit)`
- `io.Unptr[A any](ptra *A) IO[A]` - retrieves the value at pointer. Fails if nil
- `io.Wrapf[A any](io IO[A], format string, args...interface{}) IO[A]` - wraps an error with additional context information
- `io.Finally[A any](io IO[A], finalizer IO[fun.Unit]) IO[A]` - Finally runs the finalizer regardless of the success of the IO. In case finalizer fails as well, the second error is printed to log.

To and from `GoResult` - allows to handle both good value and an error:

- `io.FoldToGoResult[A any](io IO[A]) IO[GoResult[A]]` - FoldToGoResult converts either value or error to go result. It should never fail.
- `io.UnfoldGoResult[A any](iogr IO[GoResult[A]]) IO[A]` - UnfoldGoResult represents GoResult back to ordinary IO.

### Execution

To finally run all constructed computations one may use `UnsafeRunSync` or `ForEach`:

- `io.UnsafeRunSync[A any](ioa IO[A])`
- `io.ForEach[A any](io IO[A], cb func(a A))IO[fun.Unit]` - ForEach calls the provided callback after IO is completed.
- `io.RunSync[A any](io IO[A]) GoResult[A]` - RunSync is the same as UnsafeRunSync but returns GoResult.

## Parallel computing

Go routine is represented using the `Fiber[A]` interface:

```go
type Fiber[A any] interface {
	// Join waits for results of the fiber.
	// When fiber completes, this IO will complete and return the result.
	// After this fiber is closed, all join IOs fail immediately.
	Join() IO[A]
	// Closes the fiber and stops sending callbacks.
	// After closing, the respective go routine may complete
	// This is not Cancel, it does not send any signals to the fiber.
	// The work will still be done.
	Close() IO[fun.Unit]
	// Cancel sends cancellation signal to the Fiber.
	// If the fiber respects the signal, it'll stop.
	// Yet to be implemented.
	// Cancel() IO[Unit]
}
```

- `io.Start[A any](io IO[A]) IO[Fiber[A]]` - Start will start the IO in a separate go-routine. It'll establish a channel with callbacks, so that any number of listeners could join the returned fiber. When completed it'll start sending the results to the callbacks. The same value will be delivered to all listeners.
- `io.FireAndForget[A any](ioa IO[A]) IO[fun.Unit]` - FireAndForget runs the given IO in a go routine and ignores the result. It uses Fiber underneath.

### Execution contexts

Execution context is a low level resource for configuring how much processing power should be used for certain tasks. The executions are represented by `Runnable` type which is just a function without input/output. All interaction should be encapsulated inside it.

- `io.Runnable func()` - Runnable is a computation that performs some side effect and takes care of errors and panics. It task should never fail. In case it fails, application might run os.Exit(1).
- `io.ExecutionContext interface {}` - ExecutionContext is a resource capable of running tasks in parallel. NB! This is not a safe resource and it is not intended to be used directly:

```go
type ExecutionContext interface {
	// Start returns an IO which will return immediately when executed.
	// It'll place the runnable into this execution context.
	Start(neverFailingTask Runnable) IOUnit
	// Shutdown stops receiving new tasks. Subsequent start invocations will fail.
	Shutdown() IOUnit
}
```

There are two kinds of execution contexts - `UnboundedExecutionContext` and `BoundedExecutionContext`. Unbounded is recommended for IO-bound operations while bounded is for CPU-intensive tasks.

- `io.UnboundedExecutionContext() io.ExecutionContext` - UnboundedExecutionContext runs each task in a new go routine.
- `io.BoundedExecutionContext(size int64, queueLimit int) io.ExecutionContext` - BoundedExecutionContext creates an execution context that will execute tasks concurrently. Simultaneously there could be as many as `size` executions. If there are more tasks than could be started immediately they will be placed in a queue. If the queue is exhausted, `Start` will block until some tasks are run. The recommended queue size is 0 (all tasks are immediately sent to the execution). This provides immediate back pressure in case of starvation.

### Using channels with IO and parallel computations

- `io.ToChannel[A any](ch chan<- A)func(A)IO[fun.Unit]` - ToChannel saves the value to the channel.
- `io.MakeUnbufferedChannel[A any]() IO[chan A]` - MakeUnbufferedChannel allocates a new unbufered channel.
- `io.CloseChannel[A any](ch chan<- A) IO[fun.Unit]` - CloseChannel is an IO that closes the given channel.
- `io.ToChannelAndClose[A any](ch chan<- A)func(A)IO[fun.Unit]` - ToChannelAndClose sends the value to the channel and then closes the channel.
- `io.FromChannel[A any](ch <-chan A)IO[A]` - FromChannel reads a single value from the channel

### Running things in parallel

- `io.Parallel[A any](ios []IO[A]) IO[[]A]` - Parallel starts the given IOs in Go routines and waits for all results.
- `io.ParallelInExecutionContext[A any](ec ExecutionContext) func(ios []IO[A]) IO[[]A]` -  ParallelInExecutionContext starts the given IOs in the provided `ExecutionContext` and waits for all results.

- `io.ConcurrentlyFirst[A any](ios []IO[A]) IO[A]` - ConcurrentlyFirst - runs all IOs in parallel. Returns the very first result.

### Working with time

- `io.Sleep(d time.Duration)IO[fun.Unit]` - Sleep makes the IO sleep the specified time.
- `io.SleepA[A any](d time.Duration, value A)IO[A]` - SleepA sleeps and then returns the constant value
- `var ErrorTimeout` - an error that will be returned in case of timeout
- `io.WithTimeout[A any](d time.Duration) func(ioa IO[A]) IO[A]` - WithTimeout waits IO for completion for no longer than the provided duration. If there are no results, the IO will fail with timeout error.
- `io.Never[A any]() IO[A]` - Never is a simple IO that never returns.
- `io.Notify[A any](d time.Duration, value A, cb Callback[A]) IO[fun.Unit]` - Notify starts a separate thread that will call the given callback after the specified time.
- `io.NotifyToChannel[A any](d time.Duration, value A, ch chan A) IO[fun.Unit]` - NotifyToChannel sends message to channel after specified duration.

### Simple async operations

`type Callback[A any] func(A, error)` - is used as a notification mechanism for asyncronous communications.

- `io.Async[A any](k func(Callback[A])) IO[A]` - represents an asyncronous computation that will eventually call the callback.
- `io.StartInGoRoutineAndWaitForResult[A any](io IO[A]) IO[A]` - StartInGoRoutineAndWaitForResult - not very useful function. While it executes the IO in the go routine, the current thread is blocked.

## Stream

Stream represents a potentially infinite source of values.
Internally stream is a state machine that receives some input, updates internal state and produces some output.
Implementation is immutable, so we have to maintain the updated stream along the way.

### Construction

The following functions could be used to create a new stream:

- `stream.Empty[A any]()Stream[A]` - returns an empty stream.
- `stream.EmptyUnit() Stream[fun.Unit]` - returns an empty stream of units. It's more performant than `Empty[Unit]` because the same instance is being used.
- `stream.FromSlice[A any](as []A) Stream[A]`
- `stream.Lift[A any](a A) Stream[A]`
- `stream.LiftMany[A any](as ...A) Stream[A]`
- `stream.Generate[A any, S any](zero S, f func(s S) (S, A)) Stream[A]` - generates an infinite stream based on a generator function.
- `stream.Unfold[A any](zero A, f func(A) A) Stream[A]` - generates an infinite stream from previous values
- `stream.FromStepResult[A any](iosr io.IO[StepResult[A]]) Stream[A]` - basic definition of a stream - IO that returns value and continuation.
- `stream.Eval[A any](ioa io.IO[A]) Stream[A]` - Eval returns a stream of one value that is the result of IO.
- `stream.Fail[A any](err error) Stream[A]` - Fail returns a stream that fails immediately.

### Manipulation

Typical manipulations with a stream includes `Map`, `FlatMap`, `Filter` and some other helper functions.

- `stream.Map[A any, B any](stm Stream[A], f func(a A)B) Stream[B]`
- `stream.FlatMap[A any, B any](stm Stream[A], f func (a A) Stream[B]) Stream[B]`
- `stream.AndThen[A any](stm1 Stream[A], stm2 Stream[A]) Stream[A]`
- `stream.MapEval[A any, B any](stm Stream[A], f func(a A)io.IO[B]) Stream[B]`
- `stream.Filter[A any](stm Stream[A], f func(A)bool) Stream[A]`
- `stream.Flatten[A any](stm Stream[Stream[A]]) Stream[A]` - Flatten simplifies a stream of streams to just the stream of values by concatenating all inner streams.

Important functions that allow to implement stateful stream transformation:

- `stream.StateFlatMap[A any, B any, S any](stm Stream[A], zero S, f func(a A, s S) io.IO[fun.Pair[S, Stream[B]]]) Stream[B]` - consumes each element of the stream together with some state. The state is updated afterwards.
- `stream.StateFlatMapWithFinish[A any, B any, S any](stm Stream[A], zero S, f func(a A, s S) io.IO[fun.Pair[S, Stream[B]]], onFinish func(s S) Stream[B]) Stream[B]` - when the original stream finishes, there still might be some important state. This function invokes `onFinish` with the residual state value and appends the returned stream at the end.

### Execution

After constructing the desired pipeline, the stream needs to be executed.

- `stream.DrainAll[A any](stm Stream[A]) io.IO[fun.Unit]`
- `stream.AppendToSlice[A any](stm Stream[A], start []A) io.IO[[]A]`
- `stream.ToSlice[A any](stm Stream[A]) io.IO[[]A]`
- `stream.Head[A any](stm Stream[A]) io.IO[A]` - returns the first element if it exists. Otherwise - an error.
- `stream.Collect[A any](stm Stream[A], collector func (A) error) io.IO[fun.Unit]` - collects all element from the stream and for each element invokes the provided function.
- `stream.ForEach[A any](stm Stream[A], collector func (A)) io.IO[fun.Unit]` - invokes a simple function for each element of the stream.

### Channels

Provides a few utilities for working with channels:

- `stream.ToChannel[A any](stm Stream[A], ch chan A) io.IO[fun.Unit]` - sends all stream elements to the given channel
- `stream.FromChannel[A any](ch chan A) Stream[A]` - constructs a stream that reads from the given channel until the channel is open.
- `stream.PairOfChannelsToPipe[A any, B any](input chan A, output chan B) Pipe[A, B]` - PairOfChannelsToPipe - takes two channels that are being used to talk to some external process and convert them into a single pipe. It first starts a separate go routine that will continously run the input stream and send all it's contents to the `input` channel. The current thread is left with reading from the output channel.
- `stream.PipeToPairOfChannels[A any, B any](pipe Pipe[A, B]) io.IO[fun.Pair[chan A, chan B]]` - PipeToPairOfChannels converts a streaming pipe to a pair of channels that could be used to interact with external systems.

### Pipes and sinks

Pipe is as simple as a function that takes one stream and returns another stream.

Sink is a Pipe that returns a stream of units. That stream could be drained afterwards.

- `stream.NewSink[A any](f func(a A)) Sink[A]`
- `stream.Through[A any, B any](stm Stream[A], pipe Pipe[A, B]) Stream[B]`
- `stream.ToSink[A any](stm Stream[A], sink Sink[A]) Stream[fun.Unit]`

### Length manipulation

A few functions that can produce infinite stream (`Repeat`), cut the stream to known position (`Take`) or skip a few elements in the beginning (`Drop`).

- `stream.Repeat[A any](stm Stream[A]) Stream[A]` - infinitely repeat stream forever
- `stream.Take[A any](stm Stream[A], n int) Stream[A]`
- `stream.Drop[A any](stm Stream[A], n int) Stream[A]`
- `stream.ChunkN[A any](n int)func (sa Stream[A]) Stream[[]A]` - ChunkN groups elements by n and produces a stream of slices.

### Mangling

We sometimes want to intersperse the stream with some separators.

- `stream.AddSeparatorAfterEachElement[A any](stm Stream[A], sep A) Stream[A]`

### Parallel computing in streams

There is `io.Parallel` that allows to run a slice of IOs in parallel. It's not very convenient
when we have a lot of incoming requests that we wish to execute with a certain concurrency level
(to not exceed a receiver capacity).
In this case we can represent the tasks as ordinary `IO` and have a stream of tasks `Stream[IO[A]]`. The evaluation results could be represented as `GoResult[A]`.
We may wish to execute these tasks using a pool of workers of a given size.

- `stream.NewPool[A any](size int) io.IO[Pool[A]]` - NewPool creates an execution pool that will execute tasks concurrently. Simultaneously there could be as many as size executions.
- `stream.ThroughPool[A any](sa Stream[io.IO[A]], pool Pool[A]) Stream[io.GoResult[A]]` - ThroughPool runs a stream of tasks through the pool.

## Text processing

Reading and writing large text files line-by-line.

- `text.ReadLines(reader fio.Reader) stream.Stream[string]`
- `text.WriteLines(writer fio.Writer) stream.Sink[string]`
- `text.ReadOnlyFile(name string) resource.Resource[*os.File]` returns a resource for the file.
- `text.ReadLinesWithNonFinishedLine(reader fio.Reader) stream.Stream[string]` - ReadLinesWithLastNonFinishedLine reads text file line-by-line and returns the last line that is not terminated by `'\n'`.

## Slice utilities

Some utilities that are convenient when working with slices.

- `slice.Map[A any, B any](as []A, f func(A)B)(bs []B)`
- `slice.FlatMap[A any, B any](as []A, f func(A)[]B)(bs []B)`
- `slice.FoldLeft[A any, B any](as []A, zero B, f func(B, A)B) (res B)`
- `slice.Filter[A any](as []A, p func(a A) bool) (res []A)`
- `slice.FilterNot[A any](as []A, p func(a A) bool) (res []A)` - same as `Filter`, but inverses the predicate `p`.
- `slice.Flatten[A any](ass [][]A)(aas[]A)`
- `slice.AppendAll[A any](ass ...[]A) (aas []A)` - AppendAll concatenates all slices.
- `slice.GroupBy[A any, K comparable](as []A, f func(A)K) (res map[K][]A)` - GroupBy groups elements by a function that returns a key.
- `slice.GroupByMap[A any, K comparable, B any](as []A, f func(A) K, g func([]A) B) (res map[K]B)` - GroupByMap is a convenience function that groups and then maps the subslices.
- `slice.GroupByMapCount[A any, K comparable](as []A, f func(A) K) (res map[K]int)` GroupByMapCount for each key counts how often it is seen.
- `slice.Sliding[A any](as []A, size int, step int) (res [][]A)` - Sliding splits the provided slice into windows.  Each window will have the given size.  The first window starts from offset = 0. Each consequtive window starts at prev_offset + step. Last window might very well be shorter.
- `slice.Grouped[A any](as []A, size int) (res [][]A)` - Grouped partitions the slice into groups of the given size. Last partition might be smaller.
- `slice.Len[A any](as []A) int` Len returns the length of the slice. This is a normal function that can be passed around unlike the built-in `len`.

We can convert a slice to a set:

- `slice.ToSet[A comparable](as []A)(s Set[A])`

Where the `Set` type is defined as follows:

- `type Set[A comparable] map[A]struct{}`

And we can perform some operations with sets:

- `slice.SetSize[A comparable](s Set[A]) int` - SetSize returns the size of the set.

And some with arbitrary maps:

- `slice.MapValues[K comparable, A any, B any](m map[K]A, f func(A)B) (res map[K]B)` - MapValues converts values in the map using the provided function.

### Slices of numbers

Numbers support numerical operations. In generics this require defining an interface:

```go
// Number is a generic number interface that covers all Go number types.
type Number interface {
	int | int8 | int16 | int32 | int64 | 
	uint | uint8 | uint16 | uint32 | uint64 | 
	float32 | float64 |
	complex64 | complex128
}
```

Having this definition we now can aggregate slices of numbers:
 
- `slice.Sum[N Number](ns []N) (sum N)` - sums numbers.
