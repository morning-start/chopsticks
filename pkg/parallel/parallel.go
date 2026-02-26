package parallel

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type Task func() error

type Result struct {
	Index int
	Error error
}

type Pool struct {
	workers    int
	tasks     []Task
	results   []Result
	resultsMu sync.Mutex
	wg        sync.WaitGroup
Errors   []error
}

func NewPool(workers int) *Pool {
	return &Pool{
		workers: workers,
		tasks:   make([]Task, 0),
		results: make([]Result, 0),
	}
}

func (p *Pool) Add(task Task) {
	p.tasks = append(p.tasks, task)
}

func (p *Pool) AddFunc(fn func() error) {
	p.tasks = append(p.tasks, fn)
}

func (p *Pool) Run(ctx context.Context) error {
	if len(p.tasks) == 0 {
		return nil
	}

	if p.workers <= 0 {
		p.workers = 1
	}

	if p.workers > len(p.tasks) {
		p.workers = len(p.tasks)
	}

	taskChan := make(chan Task, len(p.tasks))
	resultChan := make(chan Result, len(p.tasks))

	for _, task := range p.tasks {
		taskChan <- task
	}
	close(taskChan)

	p.wg.Add(p.workers)
	for i := 0; i < p.workers; i++ {
		go func(workerId int) {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-taskChan:
					if !ok {
						return
					}
					err := task()
					resultChan <- Result{Error: err}
				}
			}
		}(i)
	}

	go func() {
		p.wg.Wait()
		close(resultChan)
	}()

	var hasError bool
	for result := range resultChan {
		if result.Error != nil {
			hasError = true
			p.Errors = append(p.Errors, result.Error)
		}
	}

	if hasError && len(p.tasks) > 1 {
		return fmt.Errorf("部分任务执行失败，共 %d 个错误", len(p.Errors))
	}

	return nil
}

func (p *Pool) RunAll() error {
	return p.Run(context.Background())
}

func RunParallel(tasks []Task, maxWorkers int) error {
	pool := NewPool(maxWorkers)
	for _, task := range tasks {
		pool.Add(task)
	}
	return pool.RunAll()
}

func RunParallelContext(ctx context.Context, tasks []Task, maxWorkers int) error {
	pool := NewPool(maxWorkers)
	for _, task := range tasks {
		pool.Add(task)
	}
	return pool.Run(ctx)
}

type DownloadTask struct {
	URL         string
	DestPath    string
	Checksum    string
	StartBytes int64
	EndBytes   int64
}

type DownloadResult struct {
	Task     DownloadTask
	Error    error
	BytesURL int64
}

type ParallelDownloader struct {
	workers    int
	downloads  []DownloadTask
	results    []DownloadResult
	progress   func(completed, total int)
	onComplete func(result DownloadResult)
}

func NewParallelDownloader(workers int) *ParallelDownloader {
	return &ParallelDownloader{
		workers:   workers,
		downloads: make([]DownloadTask, 0),
		results:   make([]DownloadResult, 0),
	}
}

func (p *ParallelDownloader) Add(task DownloadTask) {
	p.downloads = append(p.downloads, task)
}

func (p *ParallelDownloader) SetProgressCallback(fn func(completed, total int)) {
	p.progress = fn
}

func (p *ParallelDownloader) SetCompleteCallback(fn func(result DownloadResult)) {
	p.onComplete = fn
}

func (p *ParallelDownloader) Run(ctx context.Context) error {
	if len(p.downloads) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, p.workers)
	results := make(chan DownloadResult, len(p.downloads))
	var completed int32

	for i, task := range p.downloads {
		wg.Add(1)
		go func(idx int, t DownloadTask) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := DownloadResult{Task: t}
			result.Error = p.downloadFile(ctx, t)
			if result.Error == nil {
				atomic.AddInt64(&result.BytesURL, t.EndBytes - t.StartBytes)
			}

			results <- result

			newCompleted := atomic.AddInt32(&completed, 1)
			if p.progress != nil {
				p.progress(int(newCompleted), len(p.downloads))
			}
		}(i, task)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		p.results = append(p.results, result)
		if p.onComplete != nil {
			p.onComplete(result)
		}
	}

	return nil
}

func (p *ParallelDownloader) downloadFile(ctx context.Context, task DownloadTask) error {
	return nil
}

func (p *ParallelDownloader) Results() []DownloadResult {
	return p.results
}

func (p *ParallelDownloader) SuccessCount() int {
	count := 0
	for _, r := range p.results {
		if r.Error == nil {
			count++
		}
	}
	return count
}

func (p *ParallelDownloader) ErrorCount() int {
	count := 0
	for _, r := range p.results {
		if r.Error != nil {
			count++
		}
	}
	return count
}

type UpdateResult struct {
	Name    string
	Success bool
	Error   error
}

type ParallelUpdater struct {
	workers int
	apps    []string
	results []UpdateResult
	updateFn func(name string) error
	progress func(completed, total int)
}

func NewParallelUpdater(workers int) *ParallelUpdater {
	return &ParallelUpdater{
		workers: workers,
		apps:   make([]string, 0),
		results: make([]UpdateResult, 0),
	}
}

func (p *ParallelUpdater) AddApp(name string) {
	p.apps = append(p.apps, name)
}

func (p *ParallelUpdater) SetUpdateFunc(fn func(name string) error) {
	p.updateFn = fn
}

func (p *ParallelUpdater) SetProgressCallback(fn func(completed, total int)) {
	p.progress = fn
}

func (p *ParallelUpdater) Run(ctx context.Context) error {
	if len(p.apps) == 0 || p.updateFn == nil {
		return nil
	}

	sem := make(chan struct{}, p.workers)
	var wg sync.WaitGroup
	var completed int32

	for _, app := range p.apps {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			err := p.updateFn(name)
			result := UpdateResult{
				Name:    name,
				Success: err == nil,
				Error:   err,
			}
			p.results = append(p.results, result)

			completed := atomic.AddInt32(&completed, 1)
			if p.progress != nil {
				p.progress(int(completed), len(p.apps))
			}
		}(app)
	}

	wg.Wait()
	return nil
}

func (p *ParallelUpdater) Results() []UpdateResult {
	return p.results
}

func (p *ParallelUpdater) SuccessCount() int {
	count := 0
	for _, r := range p.results {
		if r.Success {
			count++
		}
	}
	return count
}

func (p *ParallelUpdater) ErrorCount() int {
	count := 0
	for _, r := range p.results {
		if !r.Success {
			count++
		}
	}
	return count
}
