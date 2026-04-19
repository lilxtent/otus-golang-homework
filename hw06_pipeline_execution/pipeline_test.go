package hw06pipelineexecution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v any) any) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v any) any { return v }),
		g("Multiplier (* 2)", func(v any) any { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v any) any { return v.(int) + 100 }),
		g("Stringifier", func(v any) any { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v any) any) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v any) any { return v }),
		g("Multiplier (* 2)", func(v any) any { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v any) any { return v.(int) + 100 }),
		g("Stringifier", func(v any) any { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})
}

func TestExecutePipelineDoneAlreadyClosed(t *testing.T) {
	in := make(Bi)
	done := make(Bi)
	close(done)

	stages := []Stage{
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- v.(int) + 1
				}
			}()
			return out
		},
	}

	go func() {
		for _, v := range []int{1, 2, 3} {
			in <- v
		}
		close(in)
	}()

	start := time.Now()
	result := make([]int, 0, 3)

	for v := range ExecutePipeline(in, done, stages...) {
		result = append(result, v.(int))
	}

	elapsed := time.Since(start)

	require.Len(t, result, 0)
	require.Less(t, int32(elapsed), int32(sleepPerStage)+int32(fault))
}

func TestExecutePipelineDoneAffectEvenOnLockedStages(t *testing.T) {
	in := make(Bi)
	done := make(Bi)

	var wg sync.WaitGroup

	stage := func(in In) Out {
		out := make(Bi)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(out)
			for v := range in {
				out <- v
			}
		}()
		return out
	}

	close(done)

	outputDone := make(chan struct{})

	go func() {
		//nolint: revive
		for range ExecutePipeline(in, done, stage) {
		}
		close(outputDone)
	}()

	select {
	case <-outputDone:
	case <-time.After(sleepPerStage):
		t.Fatal("pipeline did not stop after done closed")
	}

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-time.After(sleepPerStage):
		t.Fatal("stage goroutine did not exit after done closed")
	}
}
