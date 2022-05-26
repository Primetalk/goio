package stream

import (
	"fmt"

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

// ThroughPool runs a stream of tasks through the pool.
func ThroughPool[A any](sa Stream[io.IO[A]], pool Pool[A]) Stream[io.GoResult[A]] {
	return Through(sa, Pipe[io.IO[A], io.GoResult[A]](pool))
}
