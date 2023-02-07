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
- `fun.Memoize[A comparable, B any](f func(a A) B) func(A) B` - Memoize returns a function that will remember the original function in a map. It's thread safe, however, not super performant.

There are also basic data structures - Unit, Pair and Either.

- `fun.Unit` - type that has only one instance
- `fun.Unit1` - the instance of the Unit type
- `fun.Pair[A any, B any]` - type that represents both A and B.
- `fun.PairV1[A any, B any](p Pair[A, B]) A` - PairV1 returns the first element of the pair.
- `fun.PairV2[A any, B any](p Pair[A, B]) B` - PairV2 returns the second element of the pair.
- `fun.PairBoth[A any, B any](p Pair[A, B]) (A, B)` - PairBoth returns both parts of the pair.
- `fun.PairSwap[A any, B any](p Pair[A, B]) Pair[B, A]` - PairSwap returns a pair with swapped parts.

For debug purposes it's useful to convert arbitrary data to strings.

- `fun.ToString[A any](a A) string` - converts the value to string using `Sprintf` `%v`.

For compatibility with `interface {}`:

- `fun.CastAsInterface[A any](a A) interface {}` - CastAsInterface casts a value of an arbitrary type as interface {}.
- `fun.Cast[A any](i Any) (a A, err error)` - Cast converts interface {} to ordinary type A. It'a simple operation i.(A) represented as a function. In case the conversion is not possible, returns an error.
- `fun.UnsafeCast[A any](i interface {}) A` - UnsafeCast converts interface {} to ordinary type A. It'a simple operation i.(A) represented as a function. In case the conversion is not possible throws a panic.

Is there a way to obtain a value of an arbitrary type?

- `fun.Nothing[A any]() A` - This function can be used anywhere where type `A` is needed. It'll always fail if invoked at runtime.

### Predicates

Predicate is a function with a boolean result type.

```go
type Predicate[A any] func(A) bool
```

- `fun.IsEqualTo[A comparable](a A) Predicate[A]` - IsEqualTo compares two arguments for equality.
- `fun.Not[A any](p Predicate[A]) Predicate[A]` - Not negates the given predicate.

## Option

A convenient data structure `Option[A]` that provides safe mechanisms to work with a potentially empty value.

- `option.None[A any]() Option[A]` - None constructs an option without value.
- `option.Some[A any](a A) Option[A]` - Some constructs an option with value.
- `option.Map[A any, B any](oa Option[A], f func(A) B) Option[B]` - Map applies a function to the value inside option if any.
- `option.Fold[A any, B any](oa Option[A], f func(A) B, g func() B) (b B)` - Fold transforms all possible values of OptionA using two provided functions.
- `option.Filter[A any](oa Option[A], predicate func(A) bool) Option[A]` - Filter leaves the value inside option only if predicate is true.
- `option.FlatMap[A any, B any](oa Option[A], f func(A) Option[B]) Option[B]` - FlatMap converts an internal value if it is present using the provided function.
- `option.Flatten[A any](ooa Option[Option[A]]) Option[A]` - Flatten simplifies option of option to just Option[A].
- `option.Get[A any](oa Option[A]) A` - Get is an unsafe function that unwraps the value from the option.
- `option.ForEach[A any](oa Option[A], f func(A))` - ForEach runs the given function on the value if it's available.
- `option.IsDefined[A any](oa Option[A]) bool` - IsDefined checks whether the option contains a value.
- `option.IsEmpty[A any](oa Option[A]) bool` - IsEmpty checks whether the option is empty.

## Either

Data structure that models sum type. It can contain either A or B.

- `either.Either[A any, B any]` - type that represents either A or B.

For `Either` there are a few helper functions:

- `either.Left[A any, B any](a A) Either[A, B]`
- `either.Right[A any, B any](b B) Either[A, B]`
- `either.IsLeft[A any, B any](eab Either[A, B]) bool`
- `either.IsRight[A any, B any](eab Either[A, B]) bool`
- `either.Fold[A any, B any, C any](eab Either[A, B], left func(A)C, right func(B)C) C` - Fold pattern matches Either with two given pattern match handlers.
- `either.GetLeft[A any, B any](eab Either[A, B]) option.Option[A]` - GetLeft returns left if it's defined.
- `either.GetRight[A any, B any](eab Either[A, B]) option.Option[B]` - GetRight returns left if it's defined.

## IO

IO encapsulates a calculation and provides a mechanism to compose a few calculations (flat map or bind).

### Error handling

An arbitrary calculation may either return good result that one might expect, or fail with an error or even panic. In Go a recommended pattern is to represent failure explicitly in the function signature by returning both result and an error. There is a convention that says that when an error is not `nil`, the result should not be used.

While the requirement to explicitly deal with errors helps implementing robust systems a lot, it is often very verbose and it advocates the bad practice of explicit control flow via `return`:

```go
a, err = foo()
if err != nil {
	return
}
```

Here a single semantically important action (`foo`) requires 4 lines of code, a branch and a return statement.

This style is also not very friendly to function composition. If you need to pass the result of `foo` further to `bar`, you'll have to first bow to error handling ceremony.

The composition of two consequtive calculations is fundamental to programming. There is even a mathematical model that studies the properties of composition of calculations.

From error handling perspective `IO[A]` provides the following features:
- encapsulates a calculation that may return `A` or might fail;
- it never panics, all panics are wrapped into `error`s and presented for handling;
- provides convenient mechanisms for composing consequtive calculations (`io.Map`, `io.FlatMap`).

### Interaction with outer world vs simple (pure) functions/calculations.

From compiler's perspective things that happen in the program can be either ordinary pure computations or modification of some state outside of the function. Pure computation is special, because it has the following benefits:
- one can execute the same computation and receive exactly the same results;
- except obtaining the result of the computation nothing is changed elsewhere;
- it's much easier to reason about what is happening in the program made of pure computations (because nothing is happening apart from the computation itself).

The ability to understand and reason about programs is crucial to the ability of creation of somewhat complex programs.

Unfortunately all these nice and desired properties break when there are so called "side effects" - change of state, outer world interaction, ... - all things that make the computation to produce a different effect (and probably return different function results) even being called with the same arguments.

`IO[A]` provides a mechanism to arrange these side-effectful computations in such a way that it's easier to predict what is happening in the program. The main feature is the delay of actual effect execution until the late moment possible. A typical IO-based program does not perform any action until it is executed. It's often possible to construct the whole large computation for a complex program and only after that perform the execution. 

### Construction

To construct an IO one may use the following functions:

- `io.Lift[A any](a A) IO[A]` - lifts a plain value to IO
- `io.LiftFunc[A any, B any](f func(A) B) func(A) IO[B]` - LiftFunc wraps the result of function into IO.
- `io.Fail[A any](err error) IO[A]` - lifts an error to IO
- `io.FromConstantGoResult[A any](gr GoResult[A]) IO[A]` - FromConstantGoResult converts an existing GoResult value into an IO. Important! This is not for normal delayed IO execution. It cannot provide any guarantee for the moment when this go result was evaluated in the first place. This is just a combination of Lift and Fail.
- `io.Eval[A any](func () (A, error)) IO[A]` - lifts an arbitrary computation. Panics are handled and represented as errors.
- `io.FromPureEffect(f func())IO[fun.Unit]` - FromPureEffect constructs IO from the simplest function signature.
- `io.Delay[A any](f func()IO[A]) IO[A]` - represents a function as a plain IO
- `io.Fold[A any, B any](io IO[A], f func(a A)IO[B], recover func (error)IO[B]) IO[B]` - handles both happy and sad paths.
- `io.Recover[A any](io IO[A], recover func(err error)IO[A])IO[A]` - handle only sad path and recover some errors to happy path.
- `io.OnError[A any](io IO[A], onError func(err error) IO[fun.Unit]) IO[A]` - OnError executes a side effect when there is an error.
- `io.Retry[A any, S any](ioa IO[A], strategy func(s S, err error) IO[option.Option[S]], zero S) IO[A]` - Retry performs the same operation a few times based on the retry strategy.
- `io.RetryS[A any, S any](ioa IO[A], strategy func(s S, err error) IO[option.Option[S]], zero S) IO[fun.Pair[A, S]]` - RetryS performs the same operation a few times based on the retry strategy. Also returns the last state of the error-handling strategy.
- `io.RetryStrategyMaxCount(substring string) func(s int, err error) IO[option.Option[int]]` - RetryStrategyMaxCount is a strategy that retries n times immediately.

### Manipulation

The following functions could be used to manipulate computations:

- `io.FlatMap[A any, B any](ioa IO[A], f func(A)IO[B]) IO[B]`
- `io.AndThen[A any, B any](ioa IO[A], iob IO[B]) IO[B]` - AndThen runs the first IO, ignores it's result and then runs the second one.
- `io.Map[A any, B any](ioA IO[A], f func(a A) B) IO[B]`
- `io.MapErr[A any, B any](ioA IO[A], f func(a A) (B, error)) IO[B]`
- `io.MapConst[A any, B any](ioA IO[A], b B) IO[B]` - MapConst ignores the result and replaces it with the given constant.
- `io.Sequence[A any](ioas []IO[A]) (res IO[[]A])`
- `io.SequenceUnit(ious []IO[Unit]) (res IOUnit)`
- `io.Unptr[A any](ptra *A) IO[A]` - retrieves the value at pointer. Fails if nil
- `io.Wrapf[A any](io IO[A], format string, args...interface{}) IO[A]` - wraps an error with additional context information
- `io.Finally[A any](io IO[A], finalizer IO[fun.Unit]) IO[A]` - Finally runs the finalizer regardless of the success of the IO. In case finalizer fails as well, the second error is printed to log.
- `io.Ignore[A any](ioa IO[A]) IOUnit` - Ignore throws away the result of IO.
- `io.MapSlice[A any, B any](ioas IO[[]A], f func(a A) B) IO[[]B]` - MapSlice converts each element of the slice inside IO[[]A] using the provided function that cannot fail.

To and from `GoResult` - allows to handle both good value and an error:

- `io.FoldToGoResult[A any](io IO[A]) IO[GoResult[A]]` - FoldToGoResult converts either value or error to go result. It should never fail.
- `io.UnfoldGoResult[A any](iogr IO[GoResult[A]]) IO[A]` - UnfoldGoResult represents GoResult back to ordinary IO.

Sometimes there is a need to perform some sideeffectful operation on a value. This can be achieved with `Consumer[A]`.
```go
// Consumer can receive an instance of A and perform some operation on it.
type Consumer[A any] func(A) IOUnit
```

- `io.CoMap[A any, B any](ca Consumer[A], f func(b B) A) Consumer[B]` - CoMap changes the input argument of the consumer.

### Execution

To finally run all constructed computations one may use `UnsafeRunSync` or `ForEach`:

- `io.UnsafeRunSync[A any](ioa IO[A])`
- `io.ForEach[A any](io IO[A], cb func(a A))IO[fun.Unit]` - ForEach calls the provided callback after IO is completed.
- `io.RunSync[A any](io IO[A]) GoResult[A]` - RunSync is the same as UnsafeRunSync but returns GoResult.

### Auxiliary functions

- `io.Memoize[A comparable, B any](f func(a A) IO[B]) func(A) IO[B]` - Memoize returns a function that will remember the original function in a map. It's thread safe, however, not super performant.

### Implementation details

IO might be implemented in various ways. Here we implement IO using continuations. A simple step in the constructed IO program might either complete (returning a result or an error), or return a continuation - another execution of the same kind. In order to obtain result we should execute the returned function.
Continuations help avoiding deeply nested stack traces. It's a universal way to do "trampolining".

- `type Continuation[A any] func() ResultOrContinuation[A]` - Continuation represents some multistep computation. Here `ResultOrContinuation[A]` is either a final result (value or error) or another continuation.
- `io.ObtainResult[A any](c Continuation[A]) (res A, err error)` - ObtainResult executes continuation until final result is obtained. There is `io.MaxContinuationDepth` variable that allows to limit the depth of continuation executions. Default value is 1000000000000.

## Resources

Resource is a thing that could only be used inside brackets - acquire/release.
```go
type Resource[A any]
```
The only allowed way to use the resource is through `Use`:
- `resource.Use[A any, B any](res Resource[A], f func(A) io.IO[B]) io.IO[B]` - Use is a only way to access the resource instance. It guarantees that the resource instance will be closed after use regardless of the failure/success result.

`ClosableIO` is a simple resource that implements Close method:
```go
type ClosableIO interface {
	Close() io.IOUnit
}
```
- `resource.FromClosableIO[A ClosableIO](ioa io.IO[A]) Resource[A]` - FromClosableIO constructs a new resource from some value that itself supports method Close.
- `resource.BoundedExecutionContextResource(size int, queueLimit int) Resource[io.ExecutionContext]` - BoundedExecutionContextResource returns a resource that is a bounded execution context.
- `resource.Fail[A any](err error) Resource[A]` - Fail creates a resource that will fail during acquisition.

## Transaction-like resources

- `transaction.Bracket[A any, T any](acquire io.IO[T], commit func(T) io.IOUnit, rollback func(T) io.IOUnit) func(tr func(t T) io.IO[A]) io.IO[A]` - Bracket executes user computation with transactional guarantee. If user computation is successful - commit is executed. Otherwise - rollback.

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
- `io.FailedFiber[A any](err error) Fiber[A]` - FailedFiber creates a fiber that will fail on Join or Close with the given error.
- `io.JoinWithTimeout[A any](f Fiber[A], d time.Duration) IO[A]` - JoinWithTimeout joins the given fiber and waits no more than the given duration.

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
- `io.BoundedExecutionContext(size int, queueLimit int) io.ExecutionContext` - BoundedExecutionContext creates an execution context that will execute tasks concurrently. Simultaneously there could be as many as `size` executions. If there are more tasks than could be started immediately they will be placed in a queue. If the queue is exhausted, `Start` will block until some tasks are run. The recommended queue size is 0 (all tasks are immediately sent to the execution). This provides immediate back pressure in case of starvation.

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
- `io.PairSequentially[A any, B any](ioa IO[A], iob IO[B]) IO[fun.Pair[A, B]]` - PairSequentially runs two IOs sequentially and returns both results.
- `io.PairParallel[A any, B any](ioa IO[A], iob IO[B]) IO[fun.Pair[A, B]]` - PairParallel runs two IOs in parallel and returns both results.
- `io.RunAlso[A any](ioa IO[A], other IOUnit) IO[A]` - RunAlso runs the other IO in parallel, but returns only the result of the first IO.
- `io.MeasureDuration[A any](ioa IO[A]) IO[fun.Pair[A, time.Duration]]` - MeasureDuration captures the wall time that was needed to evaluate the given IO.

### Working with time

- `io.Sleep(d time.Duration)IO[fun.Unit]` - Sleep makes the IO sleep the specified time.
- `io.SleepA[A any](d time.Duration, value A)IO[A]` - SleepA sleeps and then returns the constant value
- `var ErrorTimeout` - an error that will be returned in case of timeout
- `io.WithTimeout[A any](d time.Duration) func(ioa IO[A]) IO[A]` - WithTimeout waits IO for completion for no longer than the provided duration. If there are no results, the IO will fail with timeout error.
- `io.Never[A any]() IO[A]` - Never is a simple IO that never returns.
- `io.Notify[A any](d time.Duration, value A, cb Callback[A]) IO[fun.Unit]` - Notify starts a separate thread that will call the given callback after the specified time.
- `io.NotifyToChannel[A any](d time.Duration, value A, ch chan A) IO[fun.Unit]` - NotifyToChannel sends message to channel after specified duration.
- `io.AfterTimeout[A any](duration time.Duration, ioa IO[A]) IO[A]` - AfterTimeout sleeps the given time and then starts the other IO.

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
- `stream.Wrapf[A any](stm Stream[A], format string, args ...interface{}) Stream[A]` - Wrapf wraps errors produced by this stream with additional context info.
- `stream.Nats() Stream[int]` - Nats returns an infinite stream of ints starting from 1.

### Manipulation

Typical manipulations with a stream includes `Map`, `FlatMap`, `Filter` and some other helper functions.

- `stream.Map[A any, B any](stm Stream[A], f func(a A)B) Stream[B]`
- `stream.FlatMap[A any, B any](stm Stream[A], f func (a A) Stream[B]) Stream[B]`
- `stream.AndThen[A any](stm1 Stream[A], stm2 Stream[A]) Stream[A]`
- `stream.MapEval[A any, B any](stm Stream[A], f func(a A)io.IO[B]) Stream[B]`
- `stream.SideEval[A any](stm Stream[A], iounit func(A) io.IOUnit) Stream[A]` - SideEval executes a computation for each element for it's side effect. Could be used for logging, for instance.
- `stream.Filter[A any](stm Stream[A], f func(A)bool) Stream[A]`
- `stream.FilterNot[A any](stm Stream[A], f func(A)bool) Stream[A]`
- `stream.Flatten[A any](stm Stream[Stream[A]]) Stream[A]` - Flatten simplifies a stream of streams to just the stream of values by concatenating all inner streams.
- `stream.ZipWithIndex[A any](as Stream[A]) Stream[fun.Pair[int, A]]` - ZipWithIndex prepends the index to each element.

Important functions that allow to implement stateful stream transformation:

- `stream.StateFlatMap[A any, B any, S any](stm Stream[A], zero S, f func(a A, s S) io.IO[fun.Pair[S, Stream[B]]]) Stream[B]` - consumes each element of the stream together with some state. The state is updated afterwards.
- `stream.StateFlatMapWithFinish[A any, B any, S any](stm Stream[A], zero S, f func(a A, s S) io.IO[fun.Pair[S, Stream[B]]], onFinish func(s S) Stream[B]) Stream[B]` - when the original stream finishes, there still might be some important state. This function invokes `onFinish` with the residual state value and appends the returned stream at the end.
- `stream.StateFlatMapWithFinishAndFailureHandling[A any, B any, S any](stm Stream[A], zero S, f func(a A, s S) io.IO[fun.Pair[S, Stream[B]]], onFinish func(s S) Stream[B], onFailure func(s S, err error) Stream[B]) Stream[B]` -  StateFlatMapWithFinishAndFailureHandling maintains state along the way. When the source stream finishes, it invokes onFinish with the last state. If there is an error during stream evaluation, onFailure is invoked. NB! onFinish is not invoked in case of failure.
- `stream.GroupBy[A any, K comparable](stm Stream[A], key func(A) K) Stream[fun.Pair[K, []A]]` - GroupBy collects group by a user-provided key. Whenever a new key is encountered, the previous group is emitted. When the original stream finishes, the last group is emitted.
- `stream.GroupByEval[A any, K comparable](stm Stream[A], keyIO func(A) io.IO[K]) Stream[fun.Pair[K, []A]]` - GroupByEval collects group by a user-provided key (which is evaluated as IO). Whenever a new key is encountered, the previous group is emitted. When the original stream finishes, the last group is emitted.
- `stream.FoldLeftEval[A any, B any](stm Stream[A], zero B, combine func(B, A) io.IO[B]) io.IO[B]` - FoldLeftEval aggregates stream in a more simple way than StateFlatMap.
- `stream.FoldLeft[A any, B any](stm Stream[A], zero B, combine func(B, A) B) io.IO[B]` - FoldLeft aggregates stream in a more simple way than StateFlatMap.
- `stream.ToChunks[A any](size int) func(stm Stream[A]) Stream[[]A]` - ToChunks collects incoming elements in chunks of the given size.
- `stream.ChunksResize[A any](newSize int) func(stm Stream[[]A]) Stream[[]A]` - ChunksResize rebuffers chunks to the given size.

Functions to explicitly deal with failures:

- `stream.FoldToGoResult[A any](stm Stream[A]) Stream[io.GoResult[A]]` - FoldToGoResult converts a stream into a stream of go results. All go results will be non-error except probably the last one.
- `stream.UnfoldGoResult[A any](stm Stream[io.GoResult[A]], onFailure func(err error) Stream[A]) Stream[A]` - UnfoldGoResult converts a stream of GoResults back to normal stream. On the first encounter of Error, the stream fails.
- `stream.StreamFold[A any, B any](stm Stream[A], onFinish func() io.IO[B], onValue func(a A, tail Stream[A]) io.IO[B], onEmpty func(tail Stream[A]) io.IO[B], onError func(err error) io.IO[B]) io.IO[B]` - StreamFold performs arbitrary processing of a stream's single step result.

Functions to explicitly deal with failures and stream completion:

```go
// Fields should be checked in order - If Error == nil, If !IsFinished, then Value
type StreamEvent[A any] struct {
	Error      error
	IsFinished bool // true when stream has completed
	Value      A
}
```
- `stream.ToStreamEvent[A any](stm Stream[A]) Stream[StreamEvent[A]]` - ToStreamEvent converts the given stream to a stream of StreamEvents. Each normal element will become a StreamEvent with data. On a failure or finish a single element is returned before the end of the stream.

### Execution

After constructing the desired pipeline, the stream needs to be executed.

- `stream.DrainAll[A any](stm Stream[A]) io.IO[fun.Unit]`
- `stream.AppendToSlice[A any](stm Stream[A], start []A) io.IO[[]A]`
- `stream.ToSlice[A any](stm Stream[A]) io.IO[[]A]`
- `stream.Head[A any](stm Stream[A]) io.IO[A]` - returns the first element if it exists. Otherwise - an error.
- `stream.HeadAndTail[A any](stm Stream[A]) io.IO[fun.Pair[A, Stream[A]]]` - HeadAndTail returns the very first element of the stream and the rest of the stream.
- `stream.TakeAndTail[A any](stm Stream[A], n int, prefix []A) io.IO[fun.Pair[[]A, Stream[A]]] ` - TakeAndTail collects n leading elements of the stream and returns them along with the tail of the stream. If the stream is shorter, then only available elements are returned and an emtpy stream.
- `stream.Last[A any](stm Stream[A]) io.IO[A]` - Last keeps track of the current element of the stream and returns it when the stream completes.
- `stream.Collect[A any](stm Stream[A], collector func (A) error) io.IO[fun.Unit]` - collects all element from the stream and for each element invokes the provided function.
- `stream.ForEach[A any](stm Stream[A], collector func (A)) io.IO[fun.Unit]` - invokes a simple function for each element of the stream.
- `stream.Partition[A any, C any, D any](stm Stream[A], predicate func(A) bool, trueHandler func(Stream[A]) io.IO[C], falseHandler func(Stream[A]) io.IO[D]) io.IO[fun.Pair[C, D]]` - Partition divides the stream into two that are handled independently.
- `stream.FanOut[A any, B any](stm Stream[A], handlers ...func(Stream[A]) io.IO[B]) io.IO[[]B]` - FanOut distributes the same element to all handlers.

### Channels

Provides a few utilities for working with channels:

- `stream.ToChannel[A any](stm Stream[A], ch chan<- A) io.IO[fun.Unit]` - sends all stream elements to the given channel.
- `stream.ToChannels[A any](stm Stream[A], channels ... chan<- A) io.IO[fun.Unit]` - ToChannels sends each stream element to every given channel.
- `stream.FromChannel[A any](ch chan A) Stream[A]` - constructs a stream that reads from the given channel until the channel is open.
- `stream.PairOfChannelsToPipe[A any, B any](input chan A, output chan B) Pipe[A, B]` - PairOfChannelsToPipe - takes two channels that are being used to talk to some external process and convert them into a single pipe. It first starts a separate go routine that will continously run the input stream and send all it's contents to the `input` channel. The current thread is left with reading from the output channel.
- `stream.PipeToPairOfChannels[A any, B any](pipe Pipe[A, B]) io.IO[fun.Pair[chan A, chan B]]` - PipeToPairOfChannels converts a streaming pipe to a pair of channels that could be used to interact with external systems.
- `stream.ChannelBufferPipe[A any](size int) Pipe[A, A]` - ChannelBufferPipe puts incoming values into a channel and reads them from it. This allows to decouple producer and consumer.

### Pipes and sinks

Pipe is as simple as a function that takes one stream and returns another stream.

Sink is a Pipe that returns a stream of units. That stream could be drained afterwards.

- `stream.NewSink[A any](f func(a A)) Sink[A]`
- `stream.Through[A any, B any](stm Stream[A], pipe Pipe[A, B]) Stream[B]`
- `stream.ThroughPipeEval[A any, B any](stm Stream[A], pipeIO io.IO[Pipe[A, B]]) Stream[B]` - ThroughPipeEval runs the given stream through pipe that is returned by the provided pipeIO.
- `stream.ToSink[A any](stm Stream[A], sink Sink[A]) Stream[fun.Unit]`
- `stream.ConcatPipes[A any, B any, C any](pipe1 Pipe[A, B], pipe2 Pipe[B, C]) Pipe[A, C]` - ConcatPipes connects two pipes into one.
- `stream.PrependPipeToSink[A any, B any](pipe1 Pipe[A, B], sink Sink[B]) Sink[A]` - PrependPipeToSink changes the input of a sink.

### Length manipulation

A few functions that can produce infinite stream (`Repeat`), cut the stream to known position (`Take`) or skip a few elements in the beginning (`Drop`).

- `stream.Repeat[A any](stm Stream[A]) Stream[A]` - infinitely repeat stream forever
- `stream.Take[A any](stm Stream[A], n int) Stream[A]`
- `stream.Drop[A any](stm Stream[A], n int) Stream[A]`
- `stream.ChunkN[A any](n int)func (sa Stream[A]) Stream[[]A]` - ChunkN groups elements by n and produces a stream of slices.
- `stream.TakeWhile[A any](stm Stream[A], predicate func(A) bool) Stream[A]` - TakeWhile returns the beginning of the stream such that all elements satisfy the predicate.
- `stream.DropWhile[A any](stm Stream[A], predicate func(A) bool) Stream[A]` - DropWhile removes the beginning of the stream so that the new stream starts with an element that falsifies the predicate.

### Mangling

We sometimes want to intersperse the stream with some separators.

- `stream.AddSeparatorAfterEachElement[A any](stm Stream[A], sep A) Stream[A]`

### Parallel computing in streams

There is `io.Parallel` that allows to run a slice of IOs in parallel. It's not very convenient
when we have a lot of incoming requests that we wish to execute with a certain concurrency level
(to not exceed a receiver capacity).
In this case we can represent the tasks as ordinary `IO` and have a stream of tasks `Stream[IO[A]]`. The evaluation results could be represented as `GoResult[A]`.
We may wish to execute these tasks using a pool of workers of a given size.
Pool is a pipe that takes some computations and return their results (possibly failures): `Pipe[io.IO[A], io.GoResult[A]]`.

- `stream.NewPool[A any](size int) io.IO[Pipe[io.IO[A], io.GoResult[A]]]` - NewPool creates an execution pool that will execute tasks concurrently. Simultaneously there could be as many as size executions.
- `stream.ThroughPool[A any](sa Stream[io.IO[A]], pool Pipe[io.IO[A], io.GoResult[A]]) Stream[io.GoResult[A]]` - ThroughPool runs a stream of tasks through the pool.
- `stream.NewPoolFromExecutionContext[A any](ec io.ExecutionContext, capacity int) io.IO[Pool[A]]` - NewPoolFromExecutionContext creates an execution pool that will execute tasks concurrently.
- `stream.ThroughExecutionContext[A any](sa Stream[io.IO[A]], ec io.ExecutionContext, capacity int) Stream[io.GoResult[A]]` - ThroughExecutionContext runs a stream of tasks through an ExecutionContext.
- `stream.JoinManyFibers[A any](capacity int) io.IO[Pipe[io.Fiber[A], io.GoResult[A]]]` - JoinManyFibers starts a separate go-routine for each incoming Fiber. As soon as result is ready it is sent to output. At any point in time at most capacity fibers could be waited for.
- `stream.NewUnorderedPoolFromExecutionContext[A any](ec io.ExecutionContext, capacity int) io.IO[Pipe[io.IO[A], io.GoResult[A]]]` - NewUnorderedPoolFromExecutionContext creates an execution pool that will execute tasks concurrently. Each task's result will be passed to a channel as soon as it completes. Hence, the order of results will be different from the order of tasks.
- `stream.ThroughExecutionContextUnordered[A any](sa Stream[io.IO[A]], ec io.ExecutionContext, capacity int) Stream[A]` - ThroughExecutionContext runs a stream of tasks through an ExecutionContext. The order of results is not preserved! This operation recovers GoResults. This will lead to lost of good elements after one that failed. At most `capacity - 1` number of lost elements.

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
- `slice.FoldLeft[A any, B any](as []A, zero B, f func(B, A)B) (res B)` - FoldLeft folds all values in the slice using the combination function.
- `slice.Reduce[A any](as []A, f func(A, A) A) A` - Reduce aggregates all elements pairwise. Only works for non empty slices.
- `slice.Filter[A any](as []A, p func(a A) bool) (res []A)`
- `slice.FilterNot[A any](as []A, p func(a A) bool) (res []A)` - same as `Filter`, but inverses the predicate `p`.
- `slice.Partition[A any](as []A, p fun.Predicate[A]) (resT []A, resF []A)` - Partition separates elements in as according to the predicate.
- `slice.Exists[A any](p fun.Predicate[A]) fun.Predicate[[]A]` - Exists returns a predicate on slices. The predicate is true if there is an element that satisfy the given element-wise predicate. It's false for an empty slice.
- `slice.Forall[A any](p fun.Predicate[A]) fun.Predicate[[]A]` - Forall returns a predicate on slices. The predicate is true if all elements satisfy the given element-wise predicate. It's true for an empty slice.
- `slice.Collect[A any, B any](as []A, f func(a A) option.Option[B]) (bs []B)` - Collect runs through the slice, executes the given function and only keeps good returned values.
- `slice.Count[A any](as []A, predicate fun.Predicate[A]) (cnt int)` - Count counts the number of elements that satisfy the given predicate.
- `slice.Flatten[A any](ass [][]A)(aas[]A)`
- `slice.AppendAll[A any](ass ...[]A) (aas []A)` - AppendAll concatenates all slices.
- `slice.GroupBy[A any, K comparable](as []A, f func(A)K) (res map[K][]A)` - GroupBy groups elements by a function that returns a key.
- `slice.GroupByMap[A any, K comparable, B any](as []A, f func(A) K, g func([]A) B) (res map[K]B)` - GroupByMap is a convenience function that groups and then maps the subslices.
- `slice.GroupByMapCount[A any, K comparable](as []A, f func(A) K) (res map[K]int)` GroupByMapCount for each key counts how often it is seen.
- `slice.Sliding[A any](as []A, size int, step int) (res [][]A)` - Sliding splits the provided slice into windows.  Each window will have the given size.  The first window starts from offset = 0. Each consequtive window starts at prev_offset + step. Last window might very well be shorter.
- `slice.Grouped[A any](as []A, size int) (res [][]A)` - Grouped partitions the slice into groups of the given size. Last partition might be smaller.
- `slice.Len[A any](as []A) int` Len returns the length of the slice. This is a normal function that can be passed around unlike the built-in `len`.
- `slice.ForEach[A any](as []A, f func(a A) )` - ForEach executes the given function for each element of the slice.
- `slice.ZipWith[A any, B any](as []A, bs []B) (res []fun.Pair[A, B])` - ZipWith returns a slice of pairs made of elements of the two slices. The length of the result is min of both.
- `slice.ZipWithIndex[A any](as []A) (res []fun.Pair[int, A])` - ZipWithIndex prepends the index to each element.
- `slice.IndexOf[A comparable](as []A, a A) int` - IndexOf returns the index of the first occurrence of a in the slice or -1 if not found.
- `slice.Take[A any](as []A, n int) []A` - Take returns at most n elements.
- `slice.Drop[A any](as []A, n int) []A` - Drop removes initial n elements. 

We can convert a slice to a set:

- `slice.ToSet[A comparable](as []A)(s Set[A])`

### Set utilities

The `Set` type is defined as follows:

- `type Set[A comparable] map[A]struct{}`

And we can perform some operations with sets:

- `set.Contains[A comparable](set map[A]struct{}) func (A) bool` - Contains creates a predicate that will check if an element is in this set.
- `set.SetSize[A comparable](s Set[A]) int` - SetSize returns the size of the set.

### Slices of numbers

Numbers support numerical operations. In generics this require defining an interface (in `fun` package):

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
 
- `slice.Sum[N fun.Number](ns []N) (sum N)` - sums numbers.
- `slice.Range(from, to int) (res []int)` - Range starts at `from` and progresses until `to` exclusive.
- `slice.Nats(n int) []int` - Nats return slice `[]int{1, 2, ..., n}`.

## Maps utilities

Some helper functions to deal with `map[K]V`.

- `maps.Keys[K comparable, V any](m map[K]V) (keys []K)` - Keys returns keys of the map
- `maps.Merge[K comparable, V any](m1 map[K]V, m2 map[K]V, combine func(V, V) V) (m map[K]V)` - Merge combines two maps. Function `combine` is invoked when the same key is available in both maps.
- `maps.MapKeys[K1 comparable, V any, K2 comparable](m1 map[K1]V, f func(K1) K2, combine func(V, V) V) (m2 map[K2]V)` - MapKeys converts original keys to new keys.
- `map.MapValues[K comparable, A any, B any](m map[K]A, f func(A)B) (res map[K]B)` - MapValues converts values in the map using the provided function.

## Performance considerations

There is a small benchmark of stream sum that can give some idea of what performance one might expect.


In all benchmarks the same computation (`sum([1,10000]`) is performed using 3 different mechanisms:
- `BenchmarkForSum` - a simple for-loop;
- `BenchmarkSliceSum` - a slice operation `Sum`;
- `BenchmarkStreamSum` - a stream of `int`s encapsulated in `io.IO[int]` and then `stream.Sum`.

Here is the result of a run on a computer:
```
✗ go test -benchmem -run=^$ -bench ^Benchmark ./stream
goos: linux
goarch: amd64
pkg: github.com/primetalk/goio/stream
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkStreamSum-12                 94          13969686 ns/op        10241737 B/op     310057 allocs/op
BenchmarkSliceSum-12              305767              3806 ns/op               8 B/op          1 allocs/op
BenchmarkForSum-12                375842              3145 ns/op               8 B/op          1 allocs/op
PASS
ok      github.com/primetalk/goio/stream        5.224s
```

The following conclusions could be inferred:
1. There are certain tasks that might benefit from lower-level implementation ;).
2. Slice operation is slower than `for` by ~20%.
3. Handling a single stream element takes ~1.4 mks. There are ~31 allocations per single stream element. And memory overhead is ~1024 bytes per stream element.

Hence, it seems to be easy to decide, whether stream-based approach will fit a particular application needs. If the size of a single stream element is greater than 1K and it's processing requires more than 1.4 mks, then stream-based approach won't hinder the performance much.

For example, if each element is a json structure of size 10K that is received via 1G internet connection, it's transmission would take 10 mks. So stream processing will add ~10% overhead to these numbers. These numbers might be a good boundary for consideration. If element size is greater and processing is more complex, then stream overhead becomes negligible.

As a reminder, here are some benefits of the stream processing:
1. Zero boilerplate error-handling.
2. Composable and reusable functions/modules.
3. Zero debug effort (in case of following best practices of functional programming - immutability, var-free code).
4. Constant-memory (despite allocations which are short-lived and GC-consumable).
