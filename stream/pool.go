package stream

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/primetalk/goio/io"
)

// Pool is a pipe capable of running tasks in parallel.
type Pool[A any] Pipe[io.IO[A], io.GoResult[A]]

// NewPool creates an execution pool that will execute tasks concurrently.
// Simultaneously there could be as many as size executions.
func NewPool[A any](size int) io.IO[Pool[A]] {
	return io.Pure(func() Pool[A] {
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
		return Pool[A](pool)
	})
}

// NewPoolFromExecutionContext creates an execution pool that will execute tasks concurrently.
func NewPoolFromExecutionContext[A any](ec io.ExecutionContext, capacity int) io.IO[Pool[A]] {
	return io.Pure(func() Pool[A] {
		input := make(chan io.IO[A])
		output := make(chan io.Fiber[A])
		driver := func() {
			for i := range input {
				fiberIO := io.StartInExecutionContext[A](ec)(i)
				fiberGRIO := io.FoldToGoResult(fiberIO)
				fiberGR, err1 := io.UnsafeRunSync(fiberGRIO)
				if err1 == nil {
					if fiberGR.Error == nil {
						output <- fiberGR.Value
					} else {
						output <- io.FailedFiber[A](errors.Wrapf(fiberGR.Error, "NewPoolFromExecutionContext.1"))
					}
				} else {
					output <- io.FailedFiber[A](errors.Wrapf(err1, "NewPoolFromExecutionContext.2: failed after FoldToGoResult"))
				}
			}
			close(output)
		}
		go driver()
		pool := PairOfChannelsToPipe(input, output)
		pool2 := func (sioa Stream[io.IO[A]]) Stream[io.GoResult[A]] {
			return MapEval(pool(sioa), func (fa io.Fiber[A]) io.IO[io.GoResult[A]] {
				return io.FoldToGoResult(fa.Join())
			})
		}
		return Pool[A](pool2)
	})
}

// ThroughPool runs a stream of tasks through the pool.
func ThroughPool[A any](sa Stream[io.IO[A]], pool Pool[A]) Stream[io.GoResult[A]] {
	return Through(sa, Pipe[io.IO[A], io.GoResult[A]](pool))
}

// ThroughExecutionContext runs a stream of tasks through an ExecutionContext.
func ThroughExecutionContext[A any](sa Stream[io.IO[A]], ec io.ExecutionContext, capacity int) Stream[io.GoResult[A]] {
	poolIO := NewPoolFromExecutionContext[A](ec, capacity)
	return FlatMap(Eval(poolIO), func (pool Pool[A]) Stream[io.GoResult[A]] {
		return pool(sa)
	})
}
