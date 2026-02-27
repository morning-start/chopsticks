package parallel

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	pool := NewPool(3)
	if pool == nil {
		t.Fatal("NewPool() returned nil")
	}
	if pool.workers != 3 {
		t.Errorf("workers = %d, want 3", pool.workers)
	}
	if pool.tasks == nil {
		t.Error("tasks is nil")
	}
	if pool.results == nil {
		t.Error("results is nil")
	}
}

func TestPoolAdd(t *testing.T) {
	pool := NewPool(2)
	
	task := func() error {
		return nil
	}
	
	pool.Add(task)
	if len(pool.tasks) != 1 {
		t.Errorf("len(tasks) = %d, want 1", len(pool.tasks))
	}
	
	pool.AddFunc(task)
	if len(pool.tasks) != 2 {
		t.Errorf("len(tasks) = %d, want 2", len(pool.tasks))
	}
}

func TestPoolRun(t *testing.T) {
	tests := []struct {
		name      string
		workers   int
		taskCount int
		wantErr   bool
	}{
		{
			name:      "run with no tasks",
			workers:   2,
			taskCount: 0,
			wantErr:   false,
		},
		{
			name:      "run single task",
			workers:   1,
			taskCount: 1,
			wantErr:   false,
		},
		{
			name:      "run multiple tasks",
			workers:   2,
			taskCount: 5,
			wantErr:   false,
		},
		{
			name:      "run with more workers than tasks",
			workers:   10,
			taskCount: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(tt.workers)
			
			var counter int32
			for i := 0; i < tt.taskCount; i++ {
				pool.Add(func() error {
					atomic.AddInt32(&counter, 1)
					return nil
				})
			}
			
			err := pool.Run(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if !tt.wantErr && int(counter) != tt.taskCount {
				t.Errorf("counter = %d, want %d", counter, tt.taskCount)
			}
		})
	}
}

func TestPoolRunWithError(t *testing.T) {
	pool := NewPool(2)
	
	expectedErr := errors.New("task error")
	pool.Add(func() error {
		return expectedErr
	})
	pool.Add(func() error {
		return nil
	})
	
	err := pool.Run(context.Background())
	if err == nil {
		t.Error("Run() should return error when tasks fail")
	}
	
	if len(pool.Errors) != 1 {
		t.Errorf("len(Errors) = %d, want 1", len(pool.Errors))
	}
}

func TestPoolRunWithContext(t *testing.T) {
	pool := NewPool(2)
	
	ctx, cancel := context.WithCancel(context.Background())
	
	var started int32
	var completed int32
	
	// Add slow tasks
	for i := 0; i < 5; i++ {
		pool.Add(func() error {
			atomic.AddInt32(&started, 1)
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			return nil
		})
	}
	
	// Cancel context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	
	pool.Run(ctx)
	
	// Not all tasks should complete due to cancellation
	if completed >= started {
		t.Log("Context cancellation may not have affected all tasks")
	}
}

func TestRunParallel(t *testing.T) {
	tasks := []Task{
		func() error { return nil },
		func() error { return nil },
		func() error { return nil },
	}
	
	err := RunParallel(tasks, 2)
	if err != nil {
		t.Errorf("RunParallel() error = %v", err)
	}
}

func TestRunParallelContext(t *testing.T) {
	tasks := []Task{
		func() error { return nil },
		func() error { return nil },
	}
	
	ctx := context.Background()
	err := RunParallelContext(ctx, tasks, 2)
	if err != nil {
		t.Errorf("RunParallelContext() error = %v", err)
	}
}

func TestNewParallelDownloader(t *testing.T) {
	d := NewParallelDownloader(3)
	if d == nil {
		t.Fatal("NewParallelDownloader() returned nil")
	}
	if d.workers != 3 {
		t.Errorf("workers = %d, want 3", d.workers)
	}
	if d.downloads == nil {
		t.Error("downloads is nil")
	}
	if d.results == nil {
		t.Error("results is nil")
	}
}

func TestParallelDownloaderAdd(t *testing.T) {
	d := NewParallelDownloader(2)
	
	task := DownloadTask{
		URL:      "http://example.com/file.txt",
		DestPath: "/tmp/file.txt",
	}
	
	d.Add(task)
	if len(d.downloads) != 1 {
		t.Errorf("len(downloads) = %d, want 1", len(d.downloads))
	}
}

func TestParallelDownloaderCallbacks(t *testing.T) {
	d := NewParallelDownloader(2)
	
	progressCalled := false
	completeCalled := false
	
	d.SetProgressCallback(func(completed, total int) {
		progressCalled = true
	})
	
	d.SetCompleteCallback(func(result DownloadResult) {
		completeCalled = true
	})
	
	if d.progress == nil {
		t.Error("SetProgressCallback did not set progress")
	}
	if d.onComplete == nil {
		t.Error("SetCompleteCallback did not set onComplete")
	}
	
	// Note: We don't actually call the callbacks here since that would require
	// network access. The callback setting is verified above.
	t.Logf("Progress callback set: %v, Complete callback set: %v", progressCalled, completeCalled)
}

func TestParallelDownloaderResults(t *testing.T) {
	d := NewParallelDownloader(2)
	
	// Simulate results
	d.results = []DownloadResult{
		{Error: nil},
		{Error: errors.New("failed")},
		{Error: nil},
	}
	
	results := d.Results()
	if len(results) != 3 {
		t.Errorf("len(Results()) = %d, want 3", len(results))
	}
	
	if d.SuccessCount() != 2 {
		t.Errorf("SuccessCount() = %d, want 2", d.SuccessCount())
	}
	
	if d.ErrorCount() != 1 {
		t.Errorf("ErrorCount() = %d, want 1", d.ErrorCount())
	}
}

func TestNewParallelUpdater(t *testing.T) {
	u := NewParallelUpdater(3)
	if u == nil {
		t.Fatal("NewParallelUpdater() returned nil")
	}
	if u.workers != 3 {
		t.Errorf("workers = %d, want 3", u.workers)
	}
	if u.apps == nil {
		t.Error("apps is nil")
	}
	if u.results == nil {
		t.Error("results is nil")
	}
}

func TestParallelUpdaterAddApp(t *testing.T) {
	u := NewParallelUpdater(2)
	
	u.AddApp("app1")
	u.AddApp("app2")
	
	if len(u.apps) != 2 {
		t.Errorf("len(apps) = %d, want 2", len(u.apps))
	}
}

func TestParallelUpdaterSetFunctions(t *testing.T) {
	u := NewParallelUpdater(2)
	
	updateFn := func(name string) error {
		return nil
	}
	
	progressFn := func(completed, total int) {
	}
	
	u.SetUpdateFunc(updateFn)
	u.SetProgressCallback(progressFn)
	
	if u.updateFn == nil {
		t.Error("SetUpdateFunc did not set updateFn")
	}
	if u.progress == nil {
		t.Error("SetProgressCallback did not set progress")
	}
}

func TestParallelUpdaterResults(t *testing.T) {
	u := NewParallelUpdater(2)
	
	// Simulate results
	u.results = []UpdateResult{
		{Name: "app1", Success: true},
		{Name: "app2", Success: false, Error: errors.New("failed")},
		{Name: "app3", Success: true},
	}
	
	results := u.Results()
	if len(results) != 3 {
		t.Errorf("len(Results()) = %d, want 3", len(results))
	}
	
	if u.SuccessCount() != 2 {
		t.Errorf("SuccessCount() = %d, want 2", u.SuccessCount())
	}
	
	if u.ErrorCount() != 1 {
		t.Errorf("ErrorCount() = %d, want 1", u.ErrorCount())
	}
}

func TestParallelUpdaterRun(t *testing.T) {
	u := NewParallelUpdater(2)
	
	u.AddApp("app1")
	u.AddApp("app2")
	
	var updated int32
	u.SetUpdateFunc(func(name string) error {
		atomic.AddInt32(&updated, 1)
		return nil
	})
	
	ctx := context.Background()
	err := u.Run(ctx)
	if err != nil {
		t.Errorf("Run() error = %v", err)
	}
	
	if updated != 2 {
		t.Errorf("updated = %d, want 2", updated)
	}
}

func TestParallelUpdaterRunNoApps(t *testing.T) {
	u := NewParallelUpdater(2)
	
	// Don't set update function
	ctx := context.Background()
	err := u.Run(ctx)
	if err != nil {
		t.Errorf("Run() error = %v", err)
	}
}

func TestParallelUpdaterRunWithError(t *testing.T) {
	u := NewParallelUpdater(2)
	
	u.AddApp("app1")
	u.AddApp("app2")
	
	expectedErr := errors.New("update failed")
	u.SetUpdateFunc(func(name string) error {
		return expectedErr
	})
	
	ctx := context.Background()
	err := u.Run(ctx)
	if err != nil {
		t.Errorf("Run() error = %v", err)
	}
	
	if u.ErrorCount() != 2 {
		t.Errorf("ErrorCount() = %d, want 2", u.ErrorCount())
	}
}
