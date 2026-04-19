package hw06pipelineexecution

type (
	In  = <-chan any
	Out = In
	Bi  = chan any
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	current := in
	if len(stages) == 0 {
		return wrapChannelWithDoneSignal(done, current)
	}

	for _, stage := range stages {
		current = stage(wrapChannelWithDoneSignal(done, current))
	}

	return wrapChannelWithDoneSignal(done, current)
}

func wrapChannelWithDoneSignal(done In, in In) In {
	if done == nil {
		return in
	}

	wrappedIn := make(Bi)

	go func() {
		defer close(wrappedIn)

		for {
			select {
			case <-done:
				go func() {
					//nolint: revive
					for range in {
						// drain channel to avoid lock
					}
				}()
				return
			case buffer, ok := <-in:
				if ok {
					wrappedIn <- buffer
				}
			}
		}
	}()

	return wrappedIn
}
