package flow

import (
	"context"
	"runtime"
)

// Func is common function to be handled in flow
type Func func(ctx context.Context) error

// Flow is job manager
type Flow struct {
	concurrencyLevel int
}

// WrapFunc wraps simple error func as Func
func WrapFunc(f func() error) Func {
	return func(ctx context.Context) error {
		errCh := make(chan error)
		go func() {
			errCh <- f()
		}()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errCh:
				return err
			}
		}
	}
}

// New creates a new Flow
func New() *Flow {
	return &Flow{
		concurrencyLevel: runtime.NumCPU(),
	}
}

// SetConcurrencyLevel sets concurrency level
func (f *Flow) SetConcurrencyLevel(l int) {
	f.concurrencyLevel = l
}

// Serial executes given funcs as serial processes
func (f *Flow) Serial(fs ...Func) Func {
	return func(ctx context.Context) error {
		for _, _f := range fs {
			if _f == nil {
				continue
			}
			if err := _f(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// Parallel executes given funcs as parallel processes
func (f *Flow) Parallel(fs ...Func) Func {
	return func(ctx context.Context) error {
		childCtx, cancelAll := context.WithCancel(ctx)
		defer cancelAll()

		doneCh := make(chan struct{}, len(fs))
		errCh := make(chan error, len(fs))
		funcCh := make(chan Func, len(fs))

		for i := 0; i < f.concurrencyLevel; i++ {
			go func(done chan struct{}) {
				for _f := range funcCh {
					if _f == nil {
						done <- struct{}{}
						continue
					}
					if err := _f(childCtx); err != nil {
						errCh <- err
						return
					}
					done <- struct{}{}
				}
			}(doneCh)
		}

		for i := range fs {
			funcCh <- fs[i]
		}

		close(funcCh)

		for i := 0; i < len(fs); i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-doneCh:
			case err := <-errCh:
				return err
			}
		}
		return nil
	}
}
