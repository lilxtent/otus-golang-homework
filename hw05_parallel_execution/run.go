package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	jobs := make(chan Task)
	errorsChannel := make(chan error)
	defer close(errorsChannel)
	waitGroup := sync.WaitGroup{}
	context, cancelJobs := context.WithCancel(context.Background())

	go initErrorsWatcher(cancelJobs, m, errorsChannel)
	initWorkers(context, n, jobs, errorsChannel, &waitGroup)

	executeTask(context, tasks, jobs)

	close(jobs)
	waitGroup.Wait()

	if err := context.Err(); err != nil {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func initErrorsWatcher(cancelJobs context.CancelFunc, maxErrors int, errorsChannel <-chan error) {
	if maxErrors <= 0 {
		return
	}

	errorsCounter := 0

	for range errorsChannel {
		errorsCounter++

		if errorsCounter >= maxErrors {
			cancelJobs()
		}
	}
}

func initWorkers(ctx context.Context, workersCount int, jobs <-chan Task,
	errorsChan chan<- error, waitGroup *sync.WaitGroup,
) {
	for range workersCount {
		waitGroup.Add(1)

		go initWorker(ctx, jobs, errorsChan, waitGroup)
	}
}

func initWorker(ctx context.Context, jobs <-chan Task, errorsChan chan<- error, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			if err := job(); err != nil {
				errorsChan <- err
			}
		}
	}
}

func executeTask(ctx context.Context, tasks []Task, jobs chan<- Task) {
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return
		case jobs <- task:
			continue
		}
	}
}
