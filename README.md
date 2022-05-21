# Implementation of IO using go1.18 generics

## Functions

Some general functions that are sometimes useful.

- `fun.Const[A any, B any](b B)func(A)B`
- `fun.ConstUnit[B any](b B) func(Unit)B`

- `fun.Swap[A any, B any, C any](f func(a A)func(b B)C) func(b B)func(a A)C`
- `fun.Curry[A any, B any, C any](f func(a A, b B)C) func(a A)func(b B)C`

- `fun.Unit` - type that has only one instance
- `fun.Unit1` - the instance of the Unit type

- `fun.Either[A,B]` - type that represents either A or B.

- `fun.Left[A any, B any](a A) Either[A, B]`
- `fun.Right[A any, B any](b B) Either[A, B]`
- `fun.IsLeft[A any, B any](eab Either[A, B]) bool`
- `fun.IsRight[A any, B any](eab Either[A, B]) bool`
- `fun.Fold[A any, B any, C any](eab Either[A, B], left func(A)C, right func(B)C) C` - Fold pattern matches Either with two given pattern match handlers

## IO

IO encapsulates a calculation and provides a mechanism to compose a few calculations (flat map or bind).

### Construction

To construct an IO one may use the following functions:
- `io.Lift[A](a A) IO[A]` - lifts a plain value to IO
- `io.Fail[A](err error) IO[A]` - lifts an error to IO
- `io.Eval[A](func () (A, error)) IO[A]` - lifts an arbitrary computation. Panics are handled and represented as errors.
- `io.Delay[A any](f func()IO[A]) IO[A]` - represents a function as a plain IO
- `io.Fold[A any, B any](io IO[A], f func(a A)IO[B], recover func (error)IO[B]) IO[B]` - handles both happy and sad paths.
- `io.Recover[A any](io IO[A], recover func(err error)IO[A])IO[A]` - handle only sad path and recover some errors to happy path.

### Manipulation

The following functions could be used to manipulate computations:
- `io.FlatMap[A, B](ioa IO[A], f func(A)IO[B]) IO[B]`
- `io.AndThen[A any, B any](ioa IO[A], iob IO[B]) IO[B]` - AndThen runs the first IO, ignores it's result and then runs the second one.
- `io.MapPure[A, B](ioa IO[A], f func(A)B) IO[B]`
- `io.Map[A, B](ioa IO[A], f func(A)(B, error)) IO[B]`
- `io.Sequence[A any](ioas []IO[A]) (res IO[[]A])`
- `io.SequenceUnit(ious []IO[Unit]) (res IOUnit)`
- `io.Unptr[A any](ptra *A) IO[A]` - retrieves the value at pointer. Fails if nil
- `io.Wrapf[A any](io IO[A], format string, args...interface{}) IO[A]` - wraps an error with additional context information

To and from `GoResult` - allows to handle both good value and an error:
- `io.FoldToGoResult[A any](io IO[A]) IO[GoResult[A]]` - FoldToGoResult converts either value or error to go result. It should never fail.
- `io.UnfoldGoResult[A any](iogr IO[GoResult[A]]) IO[A]` - UnfoldGoResult represents GoResult back to ordinary IO.

### Execution

To finally run all constructed computations one may use
- `io.UnsafeRunSync[A](ioa IO[A])`
- `io.ForEach[A any](io IO[A], cb func(a A))IO[fun.Unit]` - ForEach calls the provided callback after IO is completed.

## Stream

Stream represents a potentially infinite source of values.
Internally stream is a state machine that receives some input, updates internal state and produces some output.
Implementation is immutable, so we have to maintain the updated stream along the way.

### Construction
- `stream.Empty[A any]()Stream[A]`
- `stream.FromSlice[A any](as []A) Stream[A]`
- `stream.Lift[A any](a A) Stream[A]`
- `stream.LiftMany[A any](as ...A) Stream[A]`
- `stream.Generate[A any, S any](zero S, f func(s S) (S, A)) Stream[A]` - generates an infinite stream based on a generator function.
- `stream.Unfold[A any](zero A, f func(A) A) Stream[A]` - generates an infinite stream from previous values

- `stream.FromStepResult[A any](iosr io.IO[StepResult[A]]) Stream[A]` - basic definition of a stream - IO that returns value and continuation.

### Manipulation

- `stream.FlatMap[A any, B any](stm Stream[A], f func (a A) Stream[B]) Stream[B]`
- `stream.AndThen[A any](stm1 Stream[A], stm2 Stream[A]) Stream[A]`
- `stream.MapEval[A any, B any](stm Stream[A], f func(a A)io.IO[B]) Stream[B]`
- `stream.MapPure[A any, B any](stm Stream[A], f func(a A)B) Stream[B]`
- `stream.StateFlatMap[A any, B any, S any](stm Stream[A], zero S, f func (a A, s S) (S, Stream[B])) Stream[B]`
- `stream.Filter[A any](stm Stream[A], f func(A)bool) Stream[A]`

### Execution

- `stream.DrainAll[A any](stm Stream[A]) io.IO[fun.Unit]`
- `stream.AppendToSlice[A any](stm Stream[A], start []A) io.IO[[]A]`
- `stream.ToSlice[A any](stm Stream[A]) io.IO[[]A]`
- `stream.Head[A any](stm Stream[A]) io.IO[A]` - returns the first element if it exists. Otherwise - an error.

### Pipes and sinks

Pipe is as simple as a function that takes one stream and returns another stream.

Sink is a Pipe that returns a stream of units. That stream could be drained.

- `stream.NewSink[A any](f func(a A)) Sink[A]`
- `stream.Through[A any, B any](stm Stream[A], pipe Pipe[A, B]) Stream[B]`
- `stream.ToSink[A any](stm Stream[A], sink Sink[A]) Stream[fun.Unit]`

### Length manipulation

- `stream.Repeat[A any](stm Stream[A]) Stream[A]` - infinitely repeat stream forever
- `stream.Take[A any](stm Stream[A], n int) Stream[A]`
- `stream.Drop[A any](stm Stream[A], n int) Stream[A]`

### Mangling

- `stream.AddSeparatorAfterEachElement[A any](stm Stream[A], sep A) Stream[A]`

## Text processing

- `text.ReadLines(reader fio.Reader) stream.Stream[string]`
- `text.WriteLines(writer fio.Writer) stream.Sink[string]`

## Slice utilities

Some utilities that are convenient when working with slices.

```
Map[A any, B any](as []A, f func(A)B)(bs []B)
FlatMap[A any, B any](as []A, f func(A)[]B)(bs []B)
FoldLeft[A any, B any](as []A, zero B, f func(B, A)B) (res B)
Filter[A any](as []A, p func(a A) bool) (res []A)
Flatten[A any](ass [][]A)(aas[]A)

ToSet[A comparable](as []A)(s Set[A])

type Set[A comparable] map[A]struct{}
```
