# contextio

`contextio` is an experimental Go 1.18 package for lazy, context-aware effects.
It provides synchronous and asynchronous effects, cooperative cancellation,
fibers, structured timeout and race operations, and cancellation-masked resource
finalization.

The package is an isolated prototype. Its API and behavior are not
production-supported, and existing production packages do not depend on it.

## Effect model

An `Effect[A]` is an opaque value with a private representation equivalent to:

```go
func(context.Context) (A, error)
```

Constructing or composing an effect does not execute user work. `Run` is the
public synchronous execution boundary:

```go
effect := contextio.Map(
	contextio.Eval(func(ctx context.Context) (int, error) {
		return 21, ctx.Err()
	}),
	func(value int) int {
		return value * 2
	},
)

value, err := contextio.Run(context.Background(), effect)
```

All execution passes through cancellation and panic-recovery boundaries. A nil
context and a zero `Effect` fail with shared deterministic errors. When an error
is returned, the result is always the zero value of its type.

Cancellation is checked before invoking a subsequent effect or user
transformation. Cancellation is cooperative: synchronous user work must observe
its context and return before execution can terminate.

## Core operations

The package provides:

- `Succeed`, `Fail`, and `Eval` for constructing effects;
- `Defer` for lazy effect construction;
- `Map` and `MapErr` for transforming successful values;
- `FlatMap` for sequencing dependent effects;
- `Fold` for handling success and failure with effects; and
- `Sleep` for a context-aware delay.

Operations that change the result type are top-level generic functions because
Go 1.18 does not support methods with additional type parameters.

Panics crossing an effect boundary become `PanicError` values. `PanicError`
retains the recovered value but does not capture a stack trace.

## Asynchronous effects

`Async` adapts callback-based work:

```go
effect := contextio.Async(func(
	ctx context.Context,
	callback contextio.Callback[int],
) contextio.Canceler {
	go func() {
		callback(42, nil)
	}()
	return func() {
		// Request cancellation of the registered operation.
	}
})
```

Registration is lazy and runs when the effect is executed. The callback may
complete synchronously or asynchronously. Only the first callback invocation is
observed; duplicate callbacks return without blocking.

If context cancellation wins, `Async` invokes the canceler once and waits for
the callback to publish terminal completion. The context error remains the
result even if the callback later reports success. A nil canceler is a no-op.

Registration must return promptly. Registration that blocks, or asynchronous
work that never publishes completion after cancellation, can block execution
indefinitely.

## Fibers

`Start` runs an effect in one goroutine under a child context and returns a
`Fiber[A]`:

```go
fiber, err := contextio.Run(ctx, contextio.Start(effect))
if err != nil {
	return err
}

value, err := contextio.Run(ctx, fiber.Join())
```

`Join` may be executed by any number of current or future joiners. Every joiner
observes the same immutable terminal outcome.

`Cancel` is idempotent. It requests child cancellation and waits for the work
goroutine to publish its terminal outcome. If cancellation wins before normal
completion, joins return the child context error. Canceling the context used to
run `Join` stops only that join operation; it does not cancel the fiber.

## Timeout and race

`Timeout` runs an effect with a derived deadline. `Race` runs a slice of effects
and returns the first observed terminal result, whether success or failure.

Both operations provide strict structured cleanup:

1. select a terminal winner;
2. cancel every loser; and
3. wait for every loser to publish terminal completion before returning.

No result-publishing goroutine is silently detached. Consequently, a loser that
ignores cancellation can delay timeout or race completion indefinitely.

An empty race fails with `ErrEmptyRace`. No fairness is guaranteed between
competitors that become ready at the same time; winner selection follows
terminal publication observed by the shared result channel.

## Resource finalization

`Bracket` combines acquisition, use, and release:

```go
effect := contextio.Bracket(
	acquire,
	func(resource Resource) contextio.Effect[Result] {
		return use(resource)
	},
	func(resource Resource, exit contextio.ExitCase) contextio.Effect[contextio.Unit] {
		return release(resource, exit)
	},
)
```

Release does not run if acquisition fails before producing a resource. After a
successful acquisition, release runs exactly once after use success, failure,
panic, or cancellation. Nested brackets release resources in reverse
acquisition order.

Release runs with cancellation masked: context values remain available, but the
release context has no deadline, done channel, or cancellation error. This
allows cleanup to finish after use cancellation, but a blocked release can also
delay completion indefinitely.

`ExitCase` reports successful, errored, or canceled use. If use and release both
fail, `CombinedError` keeps the use failure as the primary error and exposes the
release failure as `Secondary`. Its `Unwrap` method returns the primary error.

## Current IO adapters

`FromIO` and `ToIO` provide minimal lazy adapters for the repository's current
`io.IO[A]` type:

- `FromIO` runs the current IO through its synchronous execution boundary. The
  wrapped IO does not receive a context, so cancellation cannot interrupt it and
  waits until it returns.
- `ToIO` captures an explicit context and runs the context-aware effect when the
  returned current IO is executed.

The package does not provide implicit background-context, fiber, resource,
stream, channel, pool, or execution-context adapters.

## Known limitations

- Composition uses ordinary recursive Go calls. Map and flat-map chains have
  been tested at bounded depths of 10,000 and 50,000, but the package does not
  guarantee unbounded stack safety.
- Cancellation cannot interrupt context-ignoring synchronous work.
- Strict cancellation waits can block indefinitely when asynchronous work,
  fiber work, race losers, timeout work, or release logic is not cooperative.
- Race scheduling does not provide fairness among simultaneously ready
  competitors.
- `PanicError` does not capture a stack trace.
- Adapting current IO preserves laziness and error boundaries but cannot add
  cooperative cancellation to context-free work.
- Linux and Windows behavior depends on the repository CI matrix; local
  development validation alone does not provide execution evidence for those
  systems.

## Tests and benchmarks

The package includes deterministic behavioral and race-enabled tests for core
effects, asynchronous completion, fibers, timeout, race, bracket finalization,
adapters, bounded composition depth, and manually released non-cooperative work.

Run the package tests with:

```sh
go test ./experimental/contextio
go test -race ./experimental/contextio
```

Comparative microbenchmarks are defined in `effect_bench_test.go`:

```sh
go test -run '^$' -bench . -benchmem ./experimental/contextio
```

Benchmark results are machine- and toolchain-dependent and are not used as
fixed performance thresholds.
