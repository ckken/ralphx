package parallel

import (
	"context"
	"fmt"
	"sync"
)

type FuncWorker func(ctx context.Context, job Job) (WorkerResult, error)

func (fn FuncWorker) Execute(ctx context.Context, job Job) (WorkerResult, error) {
	return fn(ctx, job)
}

type LocalScheduler struct {
	Workers int
	Worker  Worker
}

func (s LocalScheduler) RunRound(ctx context.Context, jobs []Job) ([]WorkerResult, error) {
	if len(jobs) == 0 {
		return nil, nil
	}
	if s.Workers <= 1 {
		s.Workers = 1
	}
	if s.Worker == nil {
		return nil, fmt.Errorf("parallel scheduler missing worker")
	}

	results := make([]WorkerResult, len(jobs))
	jobCh := make(chan int)
	errCh := make(chan error, len(jobs))
	var wg sync.WaitGroup

	workerCount := s.Workers
	if workerCount > len(jobs) {
		workerCount = len(jobs)
	}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobCh {
				res, err := s.Worker.Execute(ctx, jobs[idx])
				if err != nil {
					errCh <- err
				}
				results[idx] = res
			}
		}()
	}

	for i := range jobs {
		jobCh <- i
	}
	close(jobCh)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return results, err
		}
	}
	return results, nil
}
