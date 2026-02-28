package async

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTaskCategory_String(t *testing.T) {
	tests := []struct {
		category TaskCategory
		want     string
	}{
		{CategoryIO, "IO"},
		{CategoryCPU, "CPU"},
		{CategoryJS, "JS"},
		{CategoryMixed, "Mixed"},
		{TaskCategory(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.category.String()
			if got != tt.want {
				t.Errorf("TaskCategory.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleTaskProfile(t *testing.T) {
	profile := SimpleTaskProfile(CategoryIO, 5)

	if profile.Category != CategoryIO {
		t.Errorf("Category = %v, want %v", profile.Category, CategoryIO)
	}
	if profile.Priority != 5 {
		t.Errorf("Priority = %v, want 5", profile.Priority)
	}
	if profile.Resources.CPU != 1.0 {
		t.Errorf("CPU = %v, want 1.0", profile.Resources.CPU)
	}
	if !profile.Resources.Network {
		t.Error("Network should be true for IO task")
	}
}

func TestIOTaskProfile(t *testing.T) {
	profile := IOTaskProfile(8)

	if profile.Category != CategoryIO {
		t.Errorf("Category = %v, want %v", profile.Category, CategoryIO)
	}
	if profile.Priority != 8 {
		t.Errorf("Priority = %v, want 8", profile.Priority)
	}
	if profile.Resources.CPU != 0.5 {
		t.Errorf("CPU = %v, want 0.5", profile.Resources.CPU)
	}
	if profile.Resources.Memory != 32 {
		t.Errorf("Memory = %v, want 32", profile.Resources.Memory)
	}
}

func TestCPUTaskProfile(t *testing.T) {
	profile := CPUTaskProfile(3)

	if profile.Category != CategoryCPU {
		t.Errorf("Category = %v, want %v", profile.Category, CategoryCPU)
	}
	if profile.Priority != 3 {
		t.Errorf("Priority = %v, want 3", profile.Priority)
	}
	if profile.Resources.CPU != 2.0 {
		t.Errorf("CPU = %v, want 2.0", profile.Resources.CPU)
	}
	if profile.Resources.Network {
		t.Error("Network should be false for CPU task")
	}
}

func TestJSTaskProfile(t *testing.T) {
	profile := JSTaskProfile(7)

	if profile.Category != CategoryJS {
		t.Errorf("Category = %v, want %v", profile.Category, CategoryJS)
	}
	if profile.Priority != 7 {
		t.Errorf("Priority = %v, want 7", profile.Priority)
	}
	if profile.Resources.Memory != 256 {
		t.Errorf("Memory = %v, want 256", profile.Resources.Memory)
	}
}

func TestMixedTaskProfile(t *testing.T) {
	profile := MixedTaskProfile(5)

	if profile.Category != CategoryMixed {
		t.Errorf("Category = %v, want %v", profile.Category, CategoryMixed)
	}
	if profile.Resources.CPU != 1.5 {
		t.Errorf("CPU = %v, want 1.5", profile.Resources.CPU)
	}
	if !profile.Resources.Network {
		t.Error("Network should be true for Mixed task")
	}
}

func TestTaskFunc(t *testing.T) {
	executed := false
	fn := func(ctx context.Context) error {
		executed = true
		return nil
	}

	task := NewTaskFunc("test-task", SimpleTaskProfile(CategoryIO, 5), fn)

	if task.ID() != "test-task" {
		t.Errorf("ID = %v, want test-task", task.ID())
	}

	err := task.Execute(context.Background())
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !executed {
		t.Error("Task function was not executed")
	}
}

func TestTaskFunc_ExecuteError(t *testing.T) {
	expectedErr := errors.New("task error")
	fn := func(ctx context.Context) error {
		return expectedErr
	}

	task := NewTaskFunc("error-task", SimpleTaskProfile(CategoryIO, 5), fn)
	err := task.Execute(context.Background())

	if err != expectedErr {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}
}

func TestTaskFuture_Wait(t *testing.T) {
	resultChan := make(chan TaskResult, 1)
	resultChan <- TaskResult{TaskID: "test", Error: nil, Duration: time.Second}

	future := &TaskFuture{
		ResultChan: resultChan,
		TaskID:     "test",
	}

	result := future.Wait()

	if result.TaskID != "test" {
		t.Errorf("TaskID = %v, want test", result.TaskID)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
	if result.Duration != time.Second {
		t.Errorf("Duration = %v, want 1s", result.Duration)
	}
}

func TestTaskFuture_WaitTimeout(t *testing.T) {
	// Test timeout
	resultChan := make(chan TaskResult)

	future := &TaskFuture{
		ResultChan: resultChan,
		TaskID:     "timeout-test",
	}

	result, ok := future.WaitTimeout(50 * time.Millisecond)

	if ok {
		t.Error("WaitTimeout should return false on timeout")
	}
	if result.Error != context.DeadlineExceeded {
		t.Errorf("Error = %v, want DeadlineExceeded", result.Error)
	}

	// Test success
	resultChan2 := make(chan TaskResult, 1)
	resultChan2 <- TaskResult{TaskID: "success", Error: nil}

	future2 := &TaskFuture{
		ResultChan: resultChan2,
		TaskID:     "success",
	}

	result2, ok2 := future2.WaitTimeout(100 * time.Millisecond)

	if !ok2 {
		t.Error("WaitTimeout should return true on success")
	}
	if result2.Error != nil {
		t.Errorf("Error = %v, want nil", result2.Error)
	}
}
