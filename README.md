# Implementation of IO using go1.18 generics

## IO

IO encapsulates a calculation and provides a mechanism to compose a few calculations (flat map or bind).

### Construction

To construct an IO one may use the following functions:
- `io.Lift[A](a A) IO[A]` - lifts a plain value to IO
- `io.Fail[A](err error) IO[A]` - lifts an error to IO
- `io.Eval[A](func () (A, error)) IO[A]` - lifts an arbitrary computation. Panics are handled and represented as errors.

### Manipulation

The following functions could be used to manipulate computations:
- `io.FlatMap[A, B](ioa IO[A], f func(A)IO[B]) IO[B]`
- `io.MapPure[A, B](ioa IO[A], f func(A)B) IO[B]`
- `io.Map[A, B](ioa IO[A], f func(A)(B, error)) IO[B]`

### Execution

To finally run all constructed computations one may use
- `io.UnsafeRunSync[A](ioa IO[A])`

## Stream

Stream represents a potentially infinite source of values.
Internally stream is a state machine that receives some input, updates internal state and produces some output.
Implementation is immutable, so we have to maintain the updated stream along the way.

### Construction
- `stream.Empty[A any]()Stream[A]`
- `stream.FromSlice[A any](as []A) Stream[A]`
- `stream.Lift[A any](a A) Stream[A]`
- `stream.LiftMany[A any](as ...A) Stream[A]`

### Manipulation

- `stream.FlatMap[A any, B any](stm Stream[A], f func (a A) Stream[B]) Stream[B]`
- `stream.AndThen[A any](stm1 Stream[A], stm2 Stream[A]) Stream[A]`
- `stream.MapEval[A any, B any](stm Stream[A], f func(a A)io.IO[B]) Stream[B]`
- `stream.MapPure[A any, B any](stm Stream[A], f func(a A)B) Stream[B]`
- `stream.StateFlatMap[A any, B any, S any](stm Stream[A], zero S, f func (a A, s S) (S, Stream[B])) Stream[B]`

### Execution

- `stream.DrainAll[A any](stm Stream[A]) io.IO[io.Unit]`
- `stream.AppendToSlice[A any](stm Stream[A], start []A) io.IO[[]A]`
- `stream.ToSlice[A any](stm Stream[A]) io.IO[[]A]`

### Length manipulation

- `stream.Repeat[A any](stm Stream[A]) Stream[A]`
- `stream.Take[A any](stm Stream[A], n int) Stream[A]`
- `stream.Drop[A any](stm Stream[A], n int) Stream[A]`

## Text processing

- `text.ReadLines(reader fio.Reader) stream.Stream[string]`
