package contextio_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/primetalk/goio/experimental/contextio"
)

func TestTimeoutCancelsAndJoinsChild(t *testing.T) {
	cleaned := make(chan struct{})
	effect := contextio.Eval(func(ctx context.Context) (int, error) {
		<-ctx.Done()
		close(cleaned)
		return 0, ctx.Err()
	})
	value, err := contextio.Run(context.Background(), contextio.Timeout(time.Millisecond, effect))
	if value != 0 || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Timeout = (%d, %v)", value, err)
	}
	select {
	case <-cleaned:
	default:
		t.Fatal("Timeout returned before child cleanup")
	}
}

func TestTimeoutReturnsSuccessBeforeDeadline(t *testing.T) {
	value, err := contextio.Run(context.Background(), contextio.Timeout(time.Hour, contextio.Succeed(42)))
	if err != nil || value != 42 {
		t.Fatalf("successful Timeout = (%d, %v)", value, err)
	}
}

func TestRaceReturnsFirstTerminalResultAndJoinsLosers(t *testing.T) {
	winnerReady := make(chan struct{})
	finishWinner := make(chan struct{})
	loserStarted := make(chan struct{})
	loserCleaned := make(chan struct{})
	winner := contextio.Eval(func(context.Context) (int, error) {
		close(winnerReady)
		<-finishWinner
		return 42, nil
	})
	loser := contextio.Eval(func(ctx context.Context) (int, error) {
		close(loserStarted)
		<-ctx.Done()
		close(loserCleaned)
		return 0, ctx.Err()
	})
	result := runIntAsync(contextio.Race([]contextio.Effect[int]{winner, loser}))
	receiveSignal(t, winnerReady)
	receiveSignal(t, loserStarted)
	close(finishWinner)
	got := receive(t, result)
	if got.err != nil || got.value != 42 {
		t.Fatalf("Race = (%d, %v)", got.value, got.err)
	}
	select {
	case <-loserCleaned:
	default:
		t.Fatal("Race returned before loser cleanup")
	}

	wantErr := errors.New("first failure")
	value, err := contextio.Run(context.Background(), contextio.Race([]contextio.Effect[int]{contextio.Fail[int](wantErr)}))
	if value != 0 || !errors.Is(err, wantErr) {
		t.Fatalf("failure Race = (%d, %v)", value, err)
	}
}

func TestRaceParentCancellationJoinsAllChildren(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	cleaned := make(chan int, 8)
	effects := make([]contextio.Effect[int], 8)
	started := make(chan int, len(effects))
	for index := range effects {
		index := index
		effects[index] = contextio.Eval(func(ctx context.Context) (int, error) {
			started <- index
			<-ctx.Done()
			cleaned <- index
			return 0, ctx.Err()
		})
	}
	result := runIntAsyncContext(parent, contextio.Race(effects))
	for range effects {
		receive(t, started)
	}
	cancel()
	got := receive(t, result)
	if !errors.Is(got.err, context.Canceled) {
		t.Fatalf("parent-canceled Race error = %v", got.err)
	}
	for range effects {
		receive(t, cleaned)
	}
}

func TestRaceEmptyAndManyPublishers(t *testing.T) {
	if _, err := contextio.Run(context.Background(), contextio.Race[int](nil)); !errors.Is(err, contextio.ErrEmptyRace) {
		t.Fatalf("empty Race error = %v", err)
	}
	effects := make([]contextio.Effect[int], 100)
	for index := range effects {
		effects[index] = contextio.Succeed(index)
	}
	value, err := contextio.Run(context.Background(), contextio.Race(effects))
	if err != nil || value < 0 || value >= len(effects) {
		t.Fatalf("many-publisher Race = (%d, %v)", value, err)
	}
}

func TestRaceAndTimeoutWaitForNonCooperativeWork(t *testing.T) {
	for _, operation := range []struct {
		name string
		make func(contextio.Effect[int], <-chan struct{}) contextio.Effect[int]
	}{
		{name: "race", make: func(loser contextio.Effect[int], loserStarted <-chan struct{}) contextio.Effect[int] {
			winner := contextio.Eval(func(context.Context) (int, error) {
				<-loserStarted
				return 1, nil
			})
			return contextio.Race([]contextio.Effect[int]{winner, loser})
		}},
		{name: "timeout", make: func(loser contextio.Effect[int], _ <-chan struct{}) contextio.Effect[int] {
			return contextio.Timeout(time.Millisecond, loser)
		}},
	} {
		t.Run(operation.name, func(t *testing.T) {
			started := make(chan struct{})
			release := make(chan struct{})
			loser := contextio.Eval(func(context.Context) (int, error) {
				close(started)
				<-release
				return 2, nil
			})
			result := runIntAsync(operation.make(loser, started))
			receiveSignal(t, started)
			assertNoResult(t, result)
			close(release)
			got := receive(t, result)
			if operation.name == "race" && (got.err != nil || got.value != 1) {
				t.Fatalf("strict race = (%d, %v)", got.value, got.err)
			}
			if operation.name == "timeout" && !errors.Is(got.err, context.DeadlineExceeded) {
				t.Fatalf("strict timeout error = %v", got.err)
			}
		})
	}
}

func TestSleepCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := contextio.Run(ctx, contextio.Sleep(time.Hour))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Sleep error = %v", err)
	}
}
