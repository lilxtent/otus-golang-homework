package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := range tasksCount {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for range tasksCount {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestMIsLessOrEquealZero(t *testing.T) {
	for _, m := range []int{-1, 0} {
		t.Run("if m <= 0 then ignore errors", func(t *testing.T) {
			var tasksCount int32 = 50
			tasks := make([]Task, 0, tasksCount)

			var runTasksCount int32

			for i := range tasksCount {
				tasks = append(tasks, func() error {
					atomic.AddInt32(&runTasksCount, 1)
					return fmt.Errorf("error from task %d", i)
				})
			}

			workersCount := 2
			maxErrorsCount := m
			_ = Run(tasks, workersCount, maxErrorsCount)

			require.Equal(t, tasksCount, runTasksCount, "all tasks run")
		})
	}
}

func TestConcurrencyWithoutSleep(t *testing.T) {
	tasksCount := 50
	tasks := make([]Task, 0, tasksCount)

	var currentlyRunning atomic.Int32
	var maxConcurrent atomic.Int32
	hangingChannel := make(chan struct{})

	for range tasksCount {
		tasks = append(tasks, func() error {
			currentlyRunning.Add(1)
			defer currentlyRunning.Add(-1)
			maxCurrent := maxConcurrent.Load()
			currently := currentlyRunning.Load()

			if currently > maxCurrent {
				maxConcurrent.CompareAndSwap(maxCurrent, currently)
			}

			<-hangingChannel

			return nil
		})
	}

	done := make(chan error)
	workersCount := 5

	go func() {
		done <- Run(tasks, workersCount, 0)
	}()

	require.Eventually(t, func() bool {
		return maxConcurrent.Load() == int32(workersCount)
	}, 2*time.Second, 5*time.Millisecond)

	close(hangingChannel)

	<-done
}
