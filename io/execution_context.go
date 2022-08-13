package io

import (
	"context"
	"fmt"
	"log"

	"github.com/primetalk/goio/fun"
	"golang.org/x/sync/semaphore"
)

// Runnable is a computation that performs some side effect and takes care of errors and panics.
// It task should never fail.
// In case it fails, application might run os.Exit(1).
type Runnable func()

// ExecutionContext is a resource capable of running tasks in parallel.
// NB! This is not a safe resource and it is not intended to be used directly.
type ExecutionContext interface {
	// Start returns an IO which will return immediately when executed.
	// It'll place the runnable into this execution context.
	Start(neverFailingTask Runnable) IOUnit
	// Close stops receiving new tasks. Subsequent start invocations will fail.
	Close() IOUnit
}

type executionContextImpl struct {
	name                     string
	neverFailingTasksChannel chan<- Runnable
}

// Start returns an IO which will return immediately when executed.
// It'll place the runnable into this execution context.
func (c executionContextImpl) Start(neverFailingTask Runnable) IOUnit {
	return FromUnit(func() (err error) {
		defer fun.RecoverToErrorVar(c.name+".start", &err)
		c.neverFailingTasksChannel <- neverFailingTask
		return
	})
}

func (c executionContextImpl) Close() IOUnit {
	return FromPureEffect(func() {
		close(c.neverFailingTasksChannel)
	})
}

var globalUnboundedExecutionContext = UnboundedExecutionContext()

// UnboundedExecutionContext runs each task in a new go routine.
func UnboundedExecutionContext() ExecutionContext {
	neverFailingTasksChannel := make(chan Runnable)
	taskRunner := func() {
		for t := range neverFailingTasksChannel {
			go t()
		}
	}
	go taskRunner()
	return executionContextImpl{
		name:                     "UnboundedExecutionContext()",
		neverFailingTasksChannel: neverFailingTasksChannel,
	}
}

// BoundedExecutionContext creates an execution context that will execute tasks concurrently.
// Simultaneously there could be as many as size executions.
// If there are more tasks than could be started immediately they will be placed in a queue.
// If the queue is exhausted, Start will block until some tasks are run.
// Recommended queue size is 0.
func BoundedExecutionContext(size int, queueLimit int) ExecutionContext {
	neverFailingTasksChannel := make(chan Runnable, queueLimit)
	ec := executionContextImpl{
		name:                     fmt.Sprintf("BoundedExecutionContext(%d, %d)", size, queueLimit),
		neverFailingTasksChannel: neverFailingTasksChannel,
	}
	sem := semaphore.NewWeighted(int64(size))
	ctx := context.Background()
	taskRunner := func() {
		for taskLoopVar := range neverFailingTasksChannel {
			err1 := sem.Acquire(ctx, 1)
			if err1 == nil {
				task := taskLoopVar
				go func() {
					defer func() {
						sem.Release(1)
					}()
					task()
				}()
			} else {
				log.Fatalf("%s failure: %+v", ec.name, err1)
			}
		}
	}
	go taskRunner()
	return ec
}
