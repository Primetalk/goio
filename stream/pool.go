package stream

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"golang.org/x/sync/semaphore"
)

// // Pool is a pipe capable of running tasks in parallel.
// type Pool[A any] Pipe[io.IO[A], io.GoResult[A]]

// NewPool creates an execution pool that will execute tasks concurrently.
// Simultaneously there could be as many as size executions.
func NewPool[A any](size int) io.IO[Pipe[io.IO[A], io.GoResult[A]]] {
	return io.Pure(func() Pipe[io.IO[A], io.GoResult[A]] {
		input := make(chan io.IO[A])
		output := make(chan io.GoResult[A])
		completedExecutorIds := make(chan int)
		executor := func(id int) {
			for i := range input {
				fmt.Println("received task")
				result := io.RunSync(i)
				fmt.Println("executed task: ", result)
				output <- result
			}
			completedExecutorIds <- id
		}
		// start executors
		for i := 0; i < size; i++ {
			go executor(i)
		}
		go func() {
			for i := 0; i < size; i++ {
				id := <-completedExecutorIds
				fmt.Println("executor completed: ", id)
			}
			close(output)
		}()
		pool := PairOfChannelsToPipe(input, output)
		return pool
	})
}

// NewPoolFromExecutionContext creates an execution pool that will execute tasks concurrently.
// After the execution context a buffer is created to allow as many as `capacity`
// parallel tasks to be executed.
// This pool won't change the order of elements.
// NB! As work starts in parallel, in case of failure
// some future elements could be evaluated even after the failed element.
// Hence we use GoResult to represent evaluation results.
func NewPoolFromExecutionContext[A any](ec io.ExecutionContext, capacity int) io.IO[Pipe[io.IO[A], io.GoResult[A]]] {
	return io.Pure(func() Pipe[io.IO[A], io.GoResult[A]] {
		return func(sioa Stream[io.IO[A]]) Stream[io.GoResult[A]] {
			fibers := MapEval(sioa, io.StartInExecutionContext[A](ec))
			// return fibers
			bufferPipe := ChannelBufferPipe[io.Fiber[A]](capacity)
			fibers2 := bufferPipe(fibers)
			return Map(fibers2, io.JoinFiberAsGoResult[A])
		}
	})
}

type channelController[A any] struct {
	totalInput        int64
	totalOutput       int64
	input             chan io.Fiber[A]
	output            chan io.GoResult[A]
	done              bool
	lock              sync.Mutex
	capacitySemaphore *semaphore.Weighted
}

func newChannelController[A any](capacity int) channelController[A] {
	return channelController[A]{
		lock:              sync.Mutex{},
		capacitySemaphore: semaphore.NewWeighted(int64(capacity)),
		input:             make(chan io.Fiber[A]),
		output:            make(chan io.GoResult[A]),
	}
}
func (c *channelController[A]) incInput(ctx context.Context) (err error) {
	err = c.capacitySemaphore.Acquire(ctx, 1)
	if err == nil {
		c.lock.Lock()
		atomic.AddInt64(&c.totalInput, 1)
		c.lock.Unlock()
	}
	return
}
func (c *channelController[A]) closeOutputIfNeeded() {
	c.lock.Lock()
	if c.done {
		ti := atomic.LoadInt64(&c.totalInput)
		to := atomic.LoadInt64(&c.totalOutput)
		if c.done && ti == to {
			close(c.output)
		}
	}
	c.lock.Unlock()
}
func (c *channelController[A]) incOutput() {
	c.capacitySemaphore.Release(1)
	c.lock.Lock()
	atomic.AddInt64(&c.totalOutput, 1)
	c.lock.Unlock()
	c.closeOutputIfNeeded()
}
func (c *channelController[A]) complete() {
	c.lock.Lock()
	c.done = true
	c.lock.Unlock()
	c.closeOutputIfNeeded()
}

// JoinManyFibers starts a separate go-routine for each incoming Fiber.
// As soon as result is ready it is sent to output.
// At any point in time at most capacity fibers could be waited for.
func JoinManyFibers[A any](capacity int) io.IO[Pipe[io.Fiber[A], io.GoResult[A]]] {
	return io.Pure(func() Pipe[io.Fiber[A], io.GoResult[A]] {
		c := newChannelController[A](capacity)
		ctx := context.Background()
		driver := func() {
			for fiber := range c.input {
				err1 := c.incInput(ctx)
				if err1 == nil {
					go func(f io.Fiber[A]) {
						defer fun.RecoverToLog("stream.JoinManyFibers.func output <- ")
						res := io.JoinFiberAsGoResult(f)
						c.output <- res
						c.incOutput()
					}(fiber)
				} else {
					log.Printf("stream.JoinManyFibers.sem.Acquire: %v", err1)
					break
				}
			}
			c.complete()
		}
		go driver()
		return PairOfChannelsToPipe(c.input, c.output)
	})
}

// NewUnorderedPoolFromExecutionContext creates an execution pool
// that will execute tasks concurrently.
// Each task's result will be passed to a channel as soon as it completes.
// Hence, the order of results will be different from the order of tasks.
func NewUnorderedPoolFromExecutionContext[A any](ec io.ExecutionContext, capacity int) io.IO[Pipe[io.IO[A], io.GoResult[A]]] {
	return io.Pure(func() Pipe[io.IO[A], io.GoResult[A]] {
		return func(sioa Stream[io.IO[A]]) Stream[io.GoResult[A]] {
			fibers := MapEval(sioa, io.StartInExecutionContext[A](ec))
			pipeStream := Eval(JoinManyFibers[A](capacity))
			return FlatMap(pipeStream, func(pipe Pipe[io.Fiber[A], io.GoResult[A]]) Stream[io.GoResult[A]] {
				return pipe(fibers)
			})
		}
	})
}

// ThroughExecutionContext runs a stream of tasks through an ExecutionContext.
// NB! This operation recovers GoResults. This will lead to lost of
// good elements after one that failed. At most `capacity - 1` number of lost elements.
func ThroughExecutionContext[A any](sa Stream[io.IO[A]], ec io.ExecutionContext, capacity int) Stream[A] {
	poolIO := NewPoolFromExecutionContext[A](ec, capacity)
	grs := ThroughPipeEval(sa, poolIO)
	return UnfoldGoResult(grs, Fail[A])
}

// ThroughExecutionContextUnordered runs a stream of tasks through an ExecutionContext.
// The order of results is not preserved!
// This operation recovers GoResults. This will lead to lost of
// good elements after one that failed. At most `capacity - 1` number of lost elements.
func ThroughExecutionContextUnordered[A any](sa Stream[io.IO[A]], ec io.ExecutionContext, capacity int) Stream[A] {
	poolIO := NewUnorderedPoolFromExecutionContext[A](ec, capacity)
	grs := ThroughPipeEval(sa, poolIO)
	return UnfoldGoResult(grs, Fail[A])
}
