package contextio_test

import (
	"context"
	"errors"
	"testing"

	"github.com/primetalk/goio/experimental/contextio"
)

func TestCoreEffectsAreLazyAndComposable(t *testing.T) {
	var calls atomicInt32
	effect := contextio.Eval(func(context.Context) (int, error) {
		calls.Add(1)
		return 20, nil
	})
	effect = contextio.Map(effect, func(value int) int { return value + 1 })
	effect = contextio.FlatMap(effect, func(value int) contextio.Effect[int] {
		return contextio.Defer(func() contextio.Effect[int] {
			calls.Add(1)
			return contextio.Succeed(value * 2)
		})
	})
	if got := calls.Load(); got != 0 {
		t.Fatalf("construction performed %d calls", got)
	}

	value, err := contextio.Run(context.Background(), effect)
	if err != nil || value != 42 {
		t.Fatalf("Run() = (%d, %v), want (42, nil)", value, err)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("execution performed %d calls, want 2", got)
	}
}

func TestCoreFailureReturnsZeroAndFoldRecovers(t *testing.T) {
	wantErr := errors.New("boom")
	failed := contextio.Map(contextio.Fail[int](wantErr), func(int) string { return "not reached" })
	value, err := contextio.Run(context.Background(), failed)
	if value != "" || !errors.Is(err, wantErr) {
		t.Fatalf("Run(failed) = (%q, %v)", value, err)
	}

	recovered := contextio.Fold(
		contextio.Fail[int](wantErr),
		func(int) contextio.Effect[string] { return contextio.Succeed("success") },
		func(err error) contextio.Effect[string] { return contextio.Succeed(err.Error()) },
	)
	value, err = contextio.Run(context.Background(), recovered)
	if err != nil || value != wantErr.Error() {
		t.Fatalf("Run(recovered) = (%q, %v)", value, err)
	}
}

func TestCorePanicBoundaries(t *testing.T) {
	tests := map[string]contextio.Effect[int]{
		"eval":  contextio.Eval(func(context.Context) (int, error) { panic("eval") }),
		"defer": contextio.Defer(func() contextio.Effect[int] { panic("defer") }),
		"map":   contextio.Map(contextio.Succeed(1), func(int) int { panic("map") }),
		"map-err": contextio.MapErr(contextio.Succeed(1), func(int) (int, error) {
			panic("map-err")
		}),
		"flat-map": contextio.FlatMap(contextio.Succeed(1), func(int) contextio.Effect[int] {
			panic("flat-map")
		}),
		"fold-success": contextio.Fold(contextio.Succeed(1),
			func(int) contextio.Effect[int] { panic("fold-success") },
			func(error) contextio.Effect[int] { return contextio.Succeed(0) }),
		"fold-failure": contextio.Fold(contextio.Fail[int](errors.New("x")),
			func(int) contextio.Effect[int] { return contextio.Succeed(0) },
			func(error) contextio.Effect[int] { panic("fold-failure") }),
	}
	for name, effect := range tests {
		t.Run(name, func(t *testing.T) {
			value, err := contextio.Run(context.Background(), effect)
			var panicErr *contextio.PanicError
			if value != 0 || !errors.As(err, &panicErr) {
				t.Fatalf("Run() = (%d, %T %v), want zero and PanicError", value, err, err)
			}
		})
	}

	_, err := contextio.Run(context.Background(), contextio.Eval(func(context.Context) (int, error) {
		panic(nil)
	}))
	var panicErr *contextio.PanicError
	if !errors.As(err, &panicErr) {
		t.Fatalf("panic(nil) error = %T %v, want PanicError", err, err)
	}
}

func TestCoreContextValidationAndPropagation(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "value")
	value, err := contextio.Run(ctx, contextio.Eval(func(ctx context.Context) (string, error) {
		return ctx.Value(key{}).(string), nil
	}))
	if err != nil || value != "value" {
		t.Fatalf("context propagation = (%q, %v)", value, err)
	}
	if _, err := contextio.Run[int](nil, contextio.Succeed(1)); !errors.Is(err, contextio.ErrNilContext) {
		t.Fatalf("nil context error = %v", err)
	}
	if _, err := contextio.Run(context.Background(), contextio.Effect[int]{}); !errors.Is(err, contextio.ErrInvalidEffect) {
		t.Fatalf("zero effect error = %v", err)
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	var called atomicBool
	_, err = contextio.Run(canceled, contextio.Eval(func(context.Context) (int, error) {
		called.Store(true)
		return 1, nil
	}))
	if !errors.Is(err, context.Canceled) || called.Load() {
		t.Fatalf("pre-cancelled Run error = %v, called = %v", err, called.Load())
	}
}

func TestCancellationStopsBeforeNextUserBoundary(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var mapped atomicBool
	source := contextio.Eval(func(context.Context) (int, error) {
		cancel()
		return 1, nil
	})
	value, err := contextio.Run(ctx, contextio.Map(source, func(value int) int {
		mapped.Store(true)
		return value + 1
	}))
	if value != 0 || !errors.Is(err, context.Canceled) || mapped.Load() {
		t.Fatalf("canceled boundary = (%d, %v), mapper called = %v", value, err, mapped.Load())
	}
}

func TestCompositionDepthCharacterization(t *testing.T) {
	for _, depth := range []int{10_000, 50_000} {
		t.Run("map", func(t *testing.T) {
			effect := contextio.Succeed(0)
			for index := 0; index < depth; index++ {
				effect = contextio.Map(effect, func(value int) int { return value + 1 })
			}
			value, err := contextio.Run(context.Background(), effect)
			if err != nil || value != depth {
				t.Fatalf("depth %d = (%d, %v)", depth, value, err)
			}
		})
		t.Run("flat-map", func(t *testing.T) {
			effect := contextio.Succeed(0)
			for index := 0; index < depth; index++ {
				effect = contextio.FlatMap(effect, func(value int) contextio.Effect[int] {
					return contextio.Succeed(value + 1)
				})
			}
			value, err := contextio.Run(context.Background(), effect)
			if err != nil || value != depth {
				t.Fatalf("depth %d = (%d, %v)", depth, value, err)
			}
		})
	}
}
