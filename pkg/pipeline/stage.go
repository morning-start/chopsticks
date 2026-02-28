package pipeline

import (
	"context"
	"fmt"
	"sync"
)

// FuncStage 函数阶段实现
type FuncStage struct {
	Name_       string
	Handler     StageFunc
	Parallelism int
	BufferSize  int
	ErrorPolicy ErrorPolicy
}

// Name 返回阶段名称
func (s *FuncStage) Name() string {
	return s.Name_
}

// Process 处理数据
func (s *FuncStage) Process(ctx context.Context, input <-chan Item, output chan<- Item) error {
	if s.Parallelism <= 1 {
		return s.processSequential(ctx, input, output)
	}
	return s.processParallel(ctx, input, output)
}

// processSequential 顺序处理
func (s *FuncStage) processSequential(ctx context.Context, input <-chan Item, output chan<- Item) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-input:
			if !ok {
				return nil
			}

			if item.Error != nil {
				if s.ErrorPolicy != ContinueOnError {
					output <- item
					continue
				}
			}

			result, err := s.Handler(ctx, item)
			if err != nil {
				result.Error = err
				if s.ErrorPolicy == StopOnError {
					output <- result
					return err
				}
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case output <- result:
			}
		}
	}
}

// processParallel 并行处理
func (s *FuncStage) processParallel(ctx context.Context, input <-chan Item, output chan<- Item) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.Parallelism)

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		case item, ok := <-input:
			if !ok {
				wg.Wait()
				return nil
			}

			wg.Add(1)
			semaphore <- struct{}{}

			go func(it Item) {
				defer wg.Done()
				defer func() { <-semaphore }()

				if it.Error != nil && s.ErrorPolicy != ContinueOnError {
					select {
					case <-ctx.Done():
						return
					case output <- it:
					}
					return
				}

				result, err := s.Handler(ctx, it)
				if err != nil {
					result.Error = err
					if s.ErrorPolicy == StopOnError {
						output <- result
						return
					}
				}

				select {
				case <-ctx.Done():
					return
				case output <- result:
				}
			}(item)
		}
	}
}

// FilterStage 过滤阶段
type FilterStage struct {
	Name_  string
	Filter func(Item) bool
}

// Name 返回阶段名称
func (s *FilterStage) Name() string {
	return s.Name_
}

// Process 过滤处理
func (s *FilterStage) Process(ctx context.Context, input <-chan Item, output chan<- Item) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-input:
			if !ok {
				return nil
			}

			if s.Filter(item) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case output <- item:
				}
			}
		}
	}
}

// MapStage 映射阶段
type MapStage struct {
	Name_  string
	Mapper func(Item) Item
}

// Name 返回阶段名称
func (s *MapStage) Name() string {
	return s.Name_
}

// Process 映射处理
func (s *MapStage) Process(ctx context.Context, input <-chan Item, output chan<- Item) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-input:
			if !ok {
				return nil
			}

			result := s.Mapper(item)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case output <- result:
			}
		}
	}
}

// ChainStages 链式连接多个阶段
func ChainStages(stages ...Stage) Stage {
	return &chainedStage{stages: stages}
}

type chainedStage struct {
	stages []Stage
}

func (s *chainedStage) Name() string {
	if len(s.stages) == 0 {
		return "empty"
	}
	return fmt.Sprintf("chain[%s->...->%s]", s.stages[0].Name(), s.stages[len(s.stages)-1].Name())
}

func (s *chainedStage) Process(ctx context.Context, input <-chan Item, output chan<- Item) error {
	if len(s.stages) == 0 {
		return nil
	}

	current := input
	for i, stage := range s.stages {
		isLast := i == len(s.stages)-1

		if isLast {
			return stage.Process(ctx, current, output)
		}

		intermediate := make(chan Item)
		go func(s Stage, in <-chan Item, out chan<- Item) {
			s.Process(ctx, in, out)
			close(out)
		}(stage, current, intermediate)

		current = intermediate
	}

	return nil
}
