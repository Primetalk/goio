package stream

import (
	"fmt"

	"github.com/primetalk/goio/io"
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
			bufferPipe := BufferPipe[io.Fiber[A]](capacity)
			fibers2 := bufferPipe(fibers)
			return Map(fibers2, io.JoinFiberAsGoResult[A])
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
