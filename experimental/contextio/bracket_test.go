package contextio_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/primetalk/goio/experimental/contextio"
)

func TestBracketAcquisitionFailureDoesNotRelease(t *testing.T) {
	wantErr := errors.New("acquire")
	var releases atomicInt32
	effect := contextio.Bracket(
		contextio.Fail[int](wantErr),
		func(int) contextio.Effect[string] { return contextio.Succeed("unused") },
		func(int, contextio.ExitCase) contextio.Effect[contextio.Unit] {
			releases.Add(1)
			return contextio.Succeed(contextio.Unit{})
		},
	)
	value, err := contextio.Run(context.Background(), effect)
	if value != "" || !errors.Is(err, wantErr) || releases.Load() != 0 {
		t.Fatalf("Bracket = (%q, %v), releases %d", value, err, releases.Load())
	}
}

func TestBracketExitCasesAndExactOnceRelease(t *testing.T) {
	wantErr := errors.New("use")
	tests := []struct {
		name     string
		use      contextio.Effect[int]
		wantExit contextio.ExitCase
		wantErr  error
		panicErr bool
	}{
		{name: "success", use: contextio.Succeed(42), wantExit: contextio.ExitSucceeded},
		{name: "failure", use: contextio.Fail[int](wantErr), wantExit: contextio.ExitErrored, wantErr: wantErr},
		{name: "panic", use: contextio.Eval(func(context.Context) (int, error) { panic("use") }), wantExit: contextio.ExitErrored, panicErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var releases atomicInt32
			var gotExit contextio.ExitCase
			effect := contextio.Bracket(
				contextio.Succeed("resource"),
				func(string) contextio.Effect[int] { return test.use },
				func(_ string, exit contextio.ExitCase) contextio.Effect[contextio.Unit] {
					return contextio.Eval(func(context.Context) (contextio.Unit, error) {
						gotExit = exit
						releases.Add(1)
						return contextio.Unit{}, nil
					})
				},
			)
			value, err := contextio.Run(context.Background(), effect)
			if releases.Load() != 1 || gotExit != test.wantExit {
				t.Fatalf("release count = %d, exit = %v", releases.Load(), gotExit)
			}
			if test.panicErr {
				assertPanicError(t, err)
			} else if !errors.Is(err, test.wantErr) {
				t.Fatalf("Bracket error = %v, want %v", err, test.wantErr)
			}
			if err != nil && value != 0 {
				t.Fatalf("failed Bracket value = %d", value)
			}
		})
	}
}

func TestBracketCancellationIsMaskedDuringReleaseAndPreservesValues(t *testing.T) {
	type key struct{}
	parent, cancel := context.WithCancel(context.WithValue(context.Background(), key{}, "kept"))
	useStarted := make(chan struct{})
	releaseObserved := make(chan struct{})
	effect := contextio.Bracket(
		contextio.Succeed(1),
		func(int) contextio.Effect[int] {
			return contextio.Eval(func(ctx context.Context) (int, error) {
				close(useStarted)
				<-ctx.Done()
				return 0, ctx.Err()
			})
		},
		func(_ int, exit contextio.ExitCase) contextio.Effect[contextio.Unit] {
			return contextio.Eval(func(ctx context.Context) (contextio.Unit, error) {
				if exit != contextio.ExitCanceled || ctx.Err() != nil || ctx.Done() != nil {
					t.Errorf("masked release: exit=%v err=%v done=%v", exit, ctx.Err(), ctx.Done())
				}
				if _, ok := ctx.Deadline(); ok {
					t.Error("masked release retained deadline")
				}
				if ctx.Value(key{}) != "kept" {
					t.Errorf("masked release value = %v", ctx.Value(key{}))
				}
				close(releaseObserved)
				return contextio.Unit{}, nil
			})
		},
	)
	result := runIntAsyncContext(parent, effect)
	receiveSignal(t, useStarted)
	cancel()
	got := receive(t, result)
	receiveSignal(t, releaseObserved)
	if !errors.Is(got.err, context.Canceled) {
		t.Fatalf("canceled Bracket error = %v", got.err)
	}
}

func TestBracketCancellationAfterAcquireSkipsUseAndStillReleases(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var used atomicBool
	var released atomicBool
	effect := contextio.Bracket(
		contextio.Eval(func(context.Context) (int, error) {
			cancel()
			return 1, nil
		}),
		func(int) contextio.Effect[int] {
			used.Store(true)
			return contextio.Succeed(42)
		},
		func(_ int, exit contextio.ExitCase) contextio.Effect[contextio.Unit] {
			return contextio.Eval(func(context.Context) (contextio.Unit, error) {
				if exit != contextio.ExitCanceled {
					t.Errorf("release exit = %v, want canceled", exit)
				}
				released.Store(true)
				return contextio.Unit{}, nil
			})
		},
	)
	value, err := contextio.Run(ctx, effect)
	if value != 0 || !errors.Is(err, context.Canceled) || used.Load() || !released.Load() {
		t.Fatalf("Bracket = (%d, %v), used=%v released=%v", value, err, used.Load(), released.Load())
	}
}

func TestBracketErrorPrecedence(t *testing.T) {
	useErr := errors.New("use")
	releaseErr := errors.New("release")
	makeBracket := func(use contextio.Effect[int]) contextio.Effect[int] {
		return contextio.Bracket(
			contextio.Succeed("resource"),
			func(string) contextio.Effect[int] { return use },
			func(string, contextio.ExitCase) contextio.Effect[contextio.Unit] {
				return contextio.Fail[contextio.Unit](releaseErr)
			},
		)
	}
	value, err := contextio.Run(context.Background(), makeBracket(contextio.Succeed(42)))
	if value != 0 || !errors.Is(err, releaseErr) {
		t.Fatalf("success plus release failure = (%d, %v)", value, err)
	}

	value, err = contextio.Run(context.Background(), makeBracket(contextio.Fail[int](useErr)))
	var combined *contextio.CombinedError
	if value != 0 || !errors.As(err, &combined) || !errors.Is(err, useErr) || combined.Secondary != releaseErr {
		t.Fatalf("combined failure = (%d, %T %v)", value, err, err)
	}
}

func TestBracketCancellationAndReleaseFailurePreserveCancellation(t *testing.T) {
	releaseErr := errors.New("release")
	ctx, cancel := context.WithCancel(context.Background())
	useStarted := make(chan struct{})
	effect := contextio.Bracket(
		contextio.Succeed(1),
		func(int) contextio.Effect[int] {
			return contextio.Eval(func(ctx context.Context) (int, error) {
				close(useStarted)
				<-ctx.Done()
				return 0, ctx.Err()
			})
		},
		func(int, contextio.ExitCase) contextio.Effect[contextio.Unit] {
			return contextio.Fail[contextio.Unit](releaseErr)
		},
	)
	result := runIntAsyncContext(ctx, effect)
	receiveSignal(t, useStarted)
	cancel()
	got := receive(t, result)
	var combined *contextio.CombinedError
	if got.value != 0 || !errors.As(got.err, &combined) || !errors.Is(got.err, context.Canceled) || combined.Secondary != releaseErr {
		t.Fatalf("cancellation plus release failure = (%d, %T %v)", got.value, got.err, got.err)
	}
}

func TestBracketConstructionPanicsStillHonorCleanup(t *testing.T) {
	var releases atomicInt32
	effect := contextio.Bracket(
		contextio.Succeed(1),
		func(int) contextio.Effect[int] { panic("construct use") },
		func(int, contextio.ExitCase) contextio.Effect[contextio.Unit] {
			releases.Add(1)
			return contextio.Succeed(contextio.Unit{})
		},
	)
	_, err := contextio.Run(context.Background(), effect)
	assertPanicError(t, err)
	if releases.Load() != 1 {
		t.Fatalf("release count after use construction panic = %d", releases.Load())
	}

	effect = contextio.Bracket(
		contextio.Succeed(1),
		func(int) contextio.Effect[int] { return contextio.Succeed(2) },
		func(int, contextio.ExitCase) contextio.Effect[contextio.Unit] { panic("construct release") },
	)
	_, err = contextio.Run(context.Background(), effect)
	assertPanicError(t, err)

	var acquireReleases atomicInt32
	effect = contextio.Bracket(
		contextio.Eval(func(context.Context) (int, error) { panic("acquire") }),
		func(int) contextio.Effect[int] { return contextio.Succeed(2) },
		func(int, contextio.ExitCase) contextio.Effect[contextio.Unit] {
			acquireReleases.Add(1)
			return contextio.Succeed(contextio.Unit{})
		},
	)
	_, err = contextio.Run(context.Background(), effect)
	assertPanicError(t, err)
	if acquireReleases.Load() != 0 {
		t.Fatalf("release count after acquire panic = %d", acquireReleases.Load())
	}

	effect = contextio.Bracket(
		contextio.Succeed(1),
		func(int) contextio.Effect[int] { return contextio.Succeed(2) },
		func(int, contextio.ExitCase) contextio.Effect[contextio.Unit] {
			return contextio.Eval(func(context.Context) (contextio.Unit, error) { panic("release") })
		},
	)
	_, err = contextio.Run(context.Background(), effect)
	assertPanicError(t, err)
}

func TestNestedBracketsReleaseInReverseOrder(t *testing.T) {
	var order []string
	outer := contextio.Bracket(
		contextio.Succeed("outer"),
		func(string) contextio.Effect[int] {
			return contextio.Bracket(
				contextio.Succeed("inner"),
				func(string) contextio.Effect[int] { return contextio.Succeed(1) },
				func(string, contextio.ExitCase) contextio.Effect[contextio.Unit] {
					return contextio.Eval(func(context.Context) (contextio.Unit, error) {
						order = append(order, "inner")
						return contextio.Unit{}, nil
					})
				},
			)
		},
		func(string, contextio.ExitCase) contextio.Effect[contextio.Unit] {
			return contextio.Eval(func(context.Context) (contextio.Unit, error) {
				order = append(order, "outer")
				return contextio.Unit{}, nil
			})
		},
	)
	if _, err := contextio.Run(context.Background(), outer); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []string{"inner", "outer"}) {
		t.Fatalf("release order = %v", order)
	}
}

func TestNestedBracketCancellationReleasesInnerThenOuter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	useStarted := make(chan struct{})
	var order []string
	var orderMu sync.Mutex
	appendOrder := func(value string) contextio.Effect[contextio.Unit] {
		return contextio.Eval(func(context.Context) (contextio.Unit, error) {
			orderMu.Lock()
			order = append(order, value)
			orderMu.Unlock()
			return contextio.Unit{}, nil
		})
	}
	effect := contextio.Bracket(
		contextio.Succeed("outer"),
		func(string) contextio.Effect[int] {
			return contextio.Bracket(
				contextio.Succeed("inner"),
				func(string) contextio.Effect[int] {
					return contextio.Eval(func(ctx context.Context) (int, error) {
						close(useStarted)
						<-ctx.Done()
						return 0, ctx.Err()
					})
				},
				func(string, contextio.ExitCase) contextio.Effect[contextio.Unit] { return appendOrder("inner") },
			)
		},
		func(string, contextio.ExitCase) contextio.Effect[contextio.Unit] { return appendOrder("outer") },
	)
	result := runIntAsyncContext(ctx, effect)
	receiveSignal(t, useStarted)
	cancel()
	got := receive(t, result)
	if !errors.Is(got.err, context.Canceled) {
		t.Fatalf("nested cancellation error = %v", got.err)
	}
	orderMu.Lock()
	defer orderMu.Unlock()
	if !reflect.DeepEqual(order, []string{"inner", "outer"}) {
		t.Fatalf("release order = %v", order)
	}
}
