package contextio_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/primetalk/goio/experimental/contextio"
	"github.com/primetalk/goio/fun"
	currentio "github.com/primetalk/goio/io"
	"github.com/primetalk/goio/transaction"
)

var benchmarkInt int
var benchmarkErr error

func BenchmarkConstantSuccess(b *testing.B) {
	prototype := contextio.Succeed(42)
	current := currentio.Lift(42)
	benchmarkRuns(b, prototype, current)
}

func BenchmarkDelayedComputation(b *testing.B) {
	prototype := contextio.Eval(func(context.Context) (int, error) { return 42, nil })
	current := currentio.Eval(func() (int, error) { return 42, nil })
	benchmarkRuns(b, prototype, current)
}

func BenchmarkMapChains(b *testing.B) {
	for _, depth := range []int{10, 100, 1_000} {
		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			prototype := contextio.Succeed(0)
			current := currentio.Lift(0)
			for index := 0; index < depth; index++ {
				prototype = contextio.Map(prototype, func(value int) int { return value + 1 })
				current = currentio.Map(current, func(value int) int { return value + 1 })
			}
			benchmarkRuns(b, prototype, current)
		})
	}
}

func BenchmarkFlatMapChains(b *testing.B) {
	for _, depth := range []int{10, 100, 1_000} {
		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			prototype := contextio.Succeed(0)
			current := currentio.Lift(0)
			for index := 0; index < depth; index++ {
				prototype = contextio.FlatMap(prototype, func(value int) contextio.Effect[int] {
					return contextio.Succeed(value + 1)
				})
				current = currentio.FlatMap(current, func(value int) currentio.IO[int] {
					return currentio.Lift(value + 1)
				})
			}
			benchmarkRuns(b, prototype, current)
		})
	}
}

func BenchmarkFailurePropagation(b *testing.B) {
	wantErr := errors.New("failure")
	prototype := contextio.Fail[int](wantErr)
	current := currentio.Fail[int](wantErr)
	for index := 0; index < 100; index++ {
		prototype = contextio.Map(prototype, func(value int) int { return value + 1 })
		current = currentio.Map(current, func(value int) int { return value + 1 })
	}
	benchmarkRuns(b, prototype, current)
}

func BenchmarkConstruction(b *testing.B) {
	b.Run("contextio/map-100", func(b *testing.B) {
		b.ReportAllocs()
		for index := 0; index < b.N; index++ {
			effect := contextio.Succeed(0)
			for depth := 0; depth < 100; depth++ {
				effect = contextio.Map(effect, func(value int) int { return value + 1 })
			}
			benchmarkInt = index
		}
	})
	b.Run("current-io/map-100", func(b *testing.B) {
		b.ReportAllocs()
		for index := 0; index < b.N; index++ {
			effect := currentio.Lift(0)
			for depth := 0; depth < 100; depth++ {
				effect = currentio.Map(effect, func(value int) int { return value + 1 })
			}
			benchmarkInt = index
		}
	})
}

func BenchmarkAsyncSynchronousCompletion(b *testing.B) {
	prototype := contextio.Async(func(_ context.Context, callback contextio.Callback[int]) contextio.Canceler {
		callback(42, nil)
		return nil
	})
	current := currentio.Async(func(callback currentio.Callback[int]) { callback(42, nil) })
	benchmarkRuns(b, prototype, current)
}

func BenchmarkFiberStartJoin(b *testing.B) {
	b.Run("contextio", func(b *testing.B) {
		b.ReportAllocs()
		ctx := context.Background()
		for index := 0; index < b.N; index++ {
			fiber, err := contextio.Run(ctx, contextio.Start(contextio.Succeed(42)))
			if err == nil {
				benchmarkInt, benchmarkErr = contextio.Run(ctx, fiber.Join())
			}
		}
	})
	b.Run("current-io", func(b *testing.B) {
		b.ReportAllocs()
		for index := 0; index < b.N; index++ {
			fiber, err := currentio.UnsafeRunSync(currentio.Start(currentio.Lift(42)))
			if err == nil {
				benchmarkInt, benchmarkErr = currentio.UnsafeRunSync(fiber.Join())
			}
		}
	})
}

func BenchmarkTimeout(b *testing.B) {
	b.Run("contextio/cooperative-cancellation", func(b *testing.B) {
		b.ReportAllocs()
		effect := contextio.Eval(func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		})
		for index := 0; index < b.N; index++ {
			benchmarkInt, benchmarkErr = contextio.Run(context.Background(), contextio.Timeout(time.Nanosecond, effect))
		}
	})
	b.Run("current-io/stop-waiting", func(b *testing.B) {
		b.ReportAllocs()
		// Current IO has no cooperative cancellation. Use bounded losing work so
		// the benchmark demonstrates stop-waiting overhead without leaking a
		// permanently blocked goroutine per iteration.
		effect := currentio.WithTimeout[int](time.Nanosecond)(currentio.SleepA(time.Nanosecond, 42))
		for index := 0; index < b.N; index++ {
			benchmarkInt, benchmarkErr = currentio.UnsafeRunSync(effect)
		}
	})
}

func BenchmarkBracketSuccess(b *testing.B) {
	prototype := contextio.Bracket(
		contextio.Succeed(1),
		func(int) contextio.Effect[int] { return contextio.Succeed(42) },
		func(int, contextio.ExitCase) contextio.Effect[contextio.Unit] {
			return contextio.Succeed(contextio.Unit{})
		},
	)
	current := transaction.Bracket[int](
		currentio.Lift(1),
		func(int) currentio.IOUnit { return currentio.Lift(fun.Unit1) },
		func(int) currentio.IOUnit { return currentio.Lift(fun.Unit1) },
	)(func(int) currentio.IO[int] { return currentio.Lift(42) })
	benchmarkRuns(b, prototype, current)
}

func BenchmarkBracketFailure(b *testing.B) {
	wantErr := errors.New("failure")
	prototype := contextio.Bracket(
		contextio.Succeed(1),
		func(int) contextio.Effect[int] { return contextio.Fail[int](wantErr) },
		func(int, contextio.ExitCase) contextio.Effect[contextio.Unit] {
			return contextio.Succeed(contextio.Unit{})
		},
	)
	current := transaction.Bracket[int](
		currentio.Lift(1),
		func(int) currentio.IOUnit { return currentio.Lift(fun.Unit1) },
		func(int) currentio.IOUnit { return currentio.Lift(fun.Unit1) },
	)(func(int) currentio.IO[int] { return currentio.Fail[int](wantErr) })
	benchmarkRuns(b, prototype, current)
}

func benchmarkRuns(b *testing.B, prototype contextio.Effect[int], current currentio.IO[int]) {
	b.Helper()
	b.Run("contextio", func(b *testing.B) {
		b.ReportAllocs()
		ctx := context.Background()
		for index := 0; index < b.N; index++ {
			benchmarkInt, benchmarkErr = contextio.Run(ctx, prototype)
		}
	})
	b.Run("current-io", func(b *testing.B) {
		b.ReportAllocs()
		for index := 0; index < b.N; index++ {
			benchmarkInt, benchmarkErr = currentio.UnsafeRunSync(current)
		}
	})
}
