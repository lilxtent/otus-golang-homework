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
	waitGroup := sync.WaitGroup{}
	context, cancelJobs := context.WithCancel(context.Background())

	go initErrorsWatcher(cancelJobs, m, errorsChannel)
	initWorkers(context, n, jobs, errorsChannel, &waitGroup)

	for _, task := range tasks {
		jobs <- task
	}

	close(jobs)
	waitGroup.Wait()
	close(errorsChannel)

	if err := context.Err(); err != nil {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func initErrorsWatcher(cancelJobs context.CancelFunc, maxErrors int, errorsChannel <-chan error) {
	errorsCounter := 0

	for range errorsChannel {
		errorsCounter++

		if maxErrors <= 0 {
			continue
		}

		if errorsCounter >= maxErrors {
			cancelJobs()
		}
	}
}

func initWorkers(ctx context.Context, workersCount int, jobs <-chan Task, errorsChan chan<- error, waitGroup *sync.WaitGroup) {
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
		case job := <-jobs:
			if error := job(); error != nil {
				errorsChan <- error
			}
		}
	}
}
