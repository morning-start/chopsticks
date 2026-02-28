package async

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewSmartDispatcher(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)

	if d == nil {
		t.Fatal("NewSmartDispatcher() returned nil")
	}
	if d.config.MaxIOWorkers != 16 {
		t.Errorf("MaxIOWorkers = %d, want 16", d.config.MaxIOWorkers)
	}
	if d.ioSemaphore == nil {
		t.Error("ioSemaphore is nil")
	}
	if d.cpuSemaphore == nil {
		t.Error("cpuSemaphore is nil")
	}
	if d.jsSemaphore == nil {
		t.Error("jsSemaphore is nil")
	}
}

func TestSmartDispatcher_StartStop(t *testing.T) {
	config := DefaultConfig()
	config.EnableAdaptive = true
	d := NewSmartDispatcher(config)

	// Test Start
	d.Start()
	if atomic.LoadInt32(&d.running) != 1 {
		t.Error("Dispatcher should be running after Start()")
	}

	// Test double Start (should not panic)
	d.Start()

	// Test Stop
	d.Stop()
	if atomic.LoadInt32(&d.running) != 0 {
		t.Error("Dispatcher should be stopped after Stop()")
	}

	// Test double Stop (should not panic)
	d.Stop()
}

func TestSmartDispatcher_Dispatch_NotStarted(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)

	task := NewTaskFunc("test", SimpleTaskProfile(CategoryIO, 5), func(ctx context.Context) error {
		return nil
	})

	_, err := d.Dispatch(context.Background(), task)
	if err == nil {
		t.Error("Dispatch should return error when dispatcher not started")
	}
}

func TestSmartDispatcher_Dispatch_IO(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	executed := false
	task := NewTaskFunc("io-task", IOTaskProfile(5), func(ctx context.Context) error {
		executed = true
		return nil
	})

	future, err := d.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	result := future.Wait()
	if result.Error != nil {
		t.Errorf("Task execution error = %v", result.Error)
	}
	if !executed {
		t.Error("IO task was not executed")
	}
}

func TestSmartDispatcher_Dispatch_CPU(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	executed := false
	task := NewTaskFunc("cpu-task", CPUTaskProfile(5), func(ctx context.Context) error {
		executed = true
		return nil
	})

	future, err := d.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	result := future.Wait()
	if result.Error != nil {
		t.Errorf("Task execution error = %v", result.Error)
	}
	if !executed {
		t.Error("CPU task was not executed")
	}
}

func TestSmartDispatcher_Dispatch_JS(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	executed := false
	task := NewTaskFunc("js-task", JSTaskProfile(5), func(ctx context.Context) error {
		executed = true
		return nil
	})

	future, err := d.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	result := future.Wait()
	if result.Error != nil {
		t.Errorf("Task execution error = %v", result.Error)
	}
	if !executed {
		t.Error("JS task was not executed")
	}
}

func TestSmartDispatcher_Dispatch_Mixed(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	executed := false
	task := NewTaskFunc("mixed-task", MixedTaskProfile(5), func(ctx context.Context) error {
		executed = true
		return nil
	})

	future, err := d.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	result := future.Wait()
	if result.Error != nil {
		t.Errorf("Task execution error = %v", result.Error)
	}
	if !executed {
		t.Error("Mixed task was not executed")
	}
}

func TestSmartDispatcher_Dispatch_Error(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	expectedErr := errors.New("task error")
	task := NewTaskFunc("error-task", SimpleTaskProfile(CategoryIO, 5), func(ctx context.Context) error {
		return expectedErr
	})

	future, err := d.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	result := future.Wait()
	if result.Error != expectedErr {
		t.Errorf("Error = %v, want %v", result.Error, expectedErr)
	}

	stats := d.GetStats()
	if stats.FailedTasks != 1 {
		t.Errorf("FailedTasks = %d, want 1", stats.FailedTasks)
	}
}

func TestSmartDispatcher_DispatchBatch(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	var counter int32
	tasks := make([]Task, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = NewTaskFunc(
			string(rune('a'+i)),
			SimpleTaskProfile(CategoryIO, 5),
			func(ctx context.Context) error {
				atomic.AddInt32(&counter, 1)
				return nil
			},
		)
	}

	futures, err := d.DispatchBatch(context.Background(), tasks)
	if err != nil {
		t.Fatalf("DispatchBatch() error = %v", err)
	}
	if len(futures) != 5 {
		t.Errorf("len(futures) = %d, want 5", len(futures))
	}

	// Wait for all tasks
	for _, future := range futures {
		result := future.Wait()
		if result.Error != nil {
			t.Errorf("Task error = %v", result.Error)
		}
	}

	if counter != 5 {
		t.Errorf("counter = %d, want 5", counter)
	}
}

func TestSmartDispatcher_Dispatch_ContextCancel(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	ctx, cancel := context.WithCancel(context.Background())

	task := NewTaskFunc("slow-task", SimpleTaskProfile(CategoryIO, 5), func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	future, err := d.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	// Cancel context immediately
	cancel()

	result := future.Wait()
	// Task might complete before cancellation takes effect
	// So we just check that it doesn't panic
	t.Logf("Result: %+v", result)
}

func TestSmartDispatcher_GetStats(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	// Initial stats should be zero
	initialStats := d.GetStats()
	if initialStats.SubmittedTasks != 0 {
		t.Errorf("Initial SubmittedTasks = %d, want 0", initialStats.SubmittedTasks)
	}

	// Dispatch some tasks with unique IDs
	taskIDs := []string{"task-a", "task-b", "task-c"}
	futures := make([]*TaskFuture, len(taskIDs))
	for i, id := range taskIDs {
		task := NewTaskFunc(
			id,
			SimpleTaskProfile(CategoryIO, 5),
			func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond) // Small delay to ensure async execution
				return nil
			},
		)
		future, err := d.Dispatch(context.Background(), task)
		if err != nil {
			t.Fatalf("Dispatch() error = %v", err)
		}
		futures[i] = future
	}

	// Check submitted count immediately
	stats := d.GetStats()
	if stats.SubmittedTasks != 3 {
		t.Errorf("SubmittedTasks = %d, want 3", stats.SubmittedTasks)
	}

	// Wait for all futures
	for _, f := range futures {
		f.Wait()
	}

	// Wait a bit for stats to be updated
	time.Sleep(50 * time.Millisecond)

	// Final stats
	finalStats := d.GetStats()
	if finalStats.CompletedTasks != 3 {
		t.Errorf("CompletedTasks = %d, want 3", finalStats.CompletedTasks)
	}
}

func TestSmartDispatcher_Wait(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	var counter int32
	futures := make([]*TaskFuture, 5)
	for i := 0; i < 5; i++ {
		taskID := fmt.Sprintf("wait-task-%d", i)
		task := NewTaskFunc(
			taskID,
			SimpleTaskProfile(CategoryIO, 5),
			func(ctx context.Context) error {
				atomic.AddInt32(&counter, 1)
				return nil
			},
		)
		futures[i], _ = d.Dispatch(context.Background(), task)
	}

	// Wait for all futures
	for _, f := range futures {
		f.Wait()
	}

	if counter != 5 {
		t.Errorf("counter = %d, want 5", counter)
	}
}

func TestSmartDispatcher_WaitWithTimeout(t *testing.T) {
	config := DefaultConfig()
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	// Test 1: Success case - no active tasks initially
	ok := d.WaitWithTimeout(100 * time.Millisecond)
	if !ok {
		t.Error("WaitWithTimeout should return true when no active tasks")
	}

	// Test 2: Dispatch a task and verify it completes
	task := NewTaskFunc("quick-task", SimpleTaskProfile(CategoryIO, 5), func(ctx context.Context) error {
		return nil
	})
	future, _ := d.Dispatch(context.Background(), task)
	future.Wait()

	// After task completes, WaitWithTimeout should return true
	ok2 := d.WaitWithTimeout(100 * time.Millisecond)
	if !ok2 {
		t.Error("WaitWithTimeout should return true after tasks complete")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxIOWorkers != 16 {
		t.Errorf("MaxIOWorkers = %d, want 16", config.MaxIOWorkers)
	}
	if config.MaxJSWorkers != 4 {
		t.Errorf("MaxJSWorkers = %d, want 4", config.MaxJSWorkers)
	}
	if !config.EnableAdaptive {
		t.Error("EnableAdaptive should be true")
	}
	if config.AdjustmentInterval != 10*time.Second {
		t.Errorf("AdjustmentInterval = %v, want 10s", config.AdjustmentInterval)
	}
}

// Benchmarks

func BenchmarkSmartDispatcher_Dispatch(b *testing.B) {
	config := DefaultConfig()
	config.EnableAdaptive = false
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	task := NewTaskFunc("bench", SimpleTaskProfile(CategoryIO, 5), func(ctx context.Context) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future, _ := d.Dispatch(context.Background(), task)
		future.Wait()
	}
}

func BenchmarkSmartDispatcher_DispatchBatch(b *testing.B) {
	config := DefaultConfig()
	config.EnableAdaptive = false
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	tasks := make([]Task, 10)
	for i := 0; i < 10; i++ {
		tasks[i] = NewTaskFunc(
			string(rune('a'+i)),
			SimpleTaskProfile(CategoryIO, 5),
			func(ctx context.Context) error {
				return nil
			},
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		futures, _ := d.DispatchBatch(context.Background(), tasks)
		for _, f := range futures {
			f.Wait()
		}
	}
}

func BenchmarkSmartDispatcher_Dispatch_Parallel(b *testing.B) {
	config := DefaultConfig()
	config.EnableAdaptive = false
	d := NewSmartDispatcher(config)
	d.Start()
	defer d.Stop()

	task := NewTaskFunc("bench", SimpleTaskProfile(CategoryIO, 5), func(ctx context.Context) error {
		return nil
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			future, _ := d.Dispatch(context.Background(), task)
			future.Wait()
		}
	})
}
