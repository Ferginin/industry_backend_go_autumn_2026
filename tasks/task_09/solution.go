package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

type inMap[T any] struct {
	index int
	inVal T
}

type outMap[R any] struct {
	index int
	res   Result[R]
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if len(in) == 0 {
		return []Result[R]{}, nil
	}

	var zero R
	inChan := make(chan inMap[T])
	outChan := make(chan outMap[R], len(in))
	wg := sync.WaitGroup{}

	go func() {
		defer close(inChan)
		for i := range len(in) {
			select {
			case <-ctx.Done():
				return
			case inChan <- inMap[T]{
				index: i,
				inVal: in[i],
			}:
			}
		}
	}()

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for inMapVal := range inChan {
				res, err := fn(ctx, inMapVal.inVal)
				if err != nil {
					res = zero
				}
				outChan <- outMap[R]{
					index: inMapVal.index,
					res: Result[R]{
						Value: res,
						Err:   err,
					},
				}

			}
		}()
	}
	go func() {
		wg.Wait()
		close(outChan)
	}()

	result := make([]Result[R], len(in))
	for r := range outChan {
		result[r.index] = r.res
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return result, nil
}
