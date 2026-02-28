package metrics

import (
	"testing"
	"time"
)

func TestNewMetricsHistory(t *testing.T) {
	h := NewMetricsHistory(10)
	if h == nil {
		t.Fatal("NewMetricsHistory returned nil")
	}
	if h.maxSize != 10 {
		t.Errorf("maxSize = %d, want 10", h.maxSize)
	}
}

func TestMetricsHistory_Add(t *testing.T) {
	h := NewMetricsHistory(3)

	// 添加 5 个快照，应该只保留最后 3 个
	for i := 0; i < 5; i++ {
		h.Add(MetricsSnapshot{
			Timestamp: time.Now(),
			Metrics:   PerformanceMetrics{TotalTasks: int64(i)},
		})
		time.Sleep(1 * time.Millisecond)
	}

	all := h.GetAll()
	if len(all) != 3 {
		t.Errorf("len(all) = %d, want 3", len(all))
	}

	// 验证保留的是最后 3 个
	if all[0].Metrics.TotalTasks != 2 {
		t.Errorf("first item TotalTasks = %d, want 2", all[0].Metrics.TotalTasks)
	}
	if all[2].Metrics.TotalTasks != 4 {
		t.Errorf("last item TotalTasks = %d, want 4", all[2].Metrics.TotalTasks)
	}
}

func TestMetricsHistory_GetRecent(t *testing.T) {
	h := NewMetricsHistory(10)

	// 添加 5 个快照
	for i := 0; i < 5; i++ {
		h.Add(MetricsSnapshot{
			Timestamp: time.Now(),
			Metrics:   PerformanceMetrics{TotalTasks: int64(i)},
		})
		time.Sleep(1 * time.Millisecond)
	}

	// 获取最近 3 个
	recent := h.GetRecent(3)
	if len(recent) != 3 {
		t.Errorf("len(recent) = %d, want 3", len(recent))
	}

	// 验证顺序
	if recent[0].Metrics.TotalTasks != 2 {
		t.Errorf("first item TotalTasks = %d, want 2", recent[0].Metrics.TotalTasks)
	}
	if recent[2].Metrics.TotalTasks != 4 {
		t.Errorf("last item TotalTasks = %d, want 4", recent[2].Metrics.TotalTasks)
	}
}

func TestMetricsHistory_Clear(t *testing.T) {
	h := NewMetricsHistory(10)

	h.Add(MetricsSnapshot{
		Timestamp: time.Now(),
		Metrics:   PerformanceMetrics{TotalTasks: 1},
	})

	if len(h.GetAll()) != 1 {
		t.Error("Expected 1 item before clear")
	}

	h.Clear()

	if len(h.GetAll()) != 0 {
		t.Error("Expected 0 items after clear")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{100, "100 B"},
		{1024, "1.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1024 * 1024 * 1024 * 1024, "1.00 TB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{100 * time.Microsecond, "100 μs"},
		{500 * time.Millisecond, "500.00 ms"},
		{2 * time.Second, "2.00 s"},
		{3 * time.Minute, "3.00 min"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.d)
		if result != tt.expected {
			t.Errorf("FormatDuration(%v) = %s, want %s", tt.d, result, tt.expected)
		}
	}
}

func TestCalculateRate(t *testing.T) {
	rate := CalculateRate(100, 10*time.Second)
	expected := 10.0
	if rate != expected {
		t.Errorf("CalculateRate(100, 10s) = %f, want %f", rate, expected)
	}

	// 测试 0 持续时间
	rate = CalculateRate(100, 0)
	if rate != 0 {
		t.Errorf("CalculateRate(100, 0) = %f, want 0", rate)
	}
}

func TestCalculatePercentage(t *testing.T) {
	pct := CalculatePercentage(25, 100)
	expected := 25.0
	if pct != expected {
		t.Errorf("CalculatePercentage(25, 100) = %f, want %f", pct, expected)
	}

	// 测试 0 总数
	pct = CalculatePercentage(25, 0)
	if pct != 0 {
		t.Errorf("CalculatePercentage(25, 0) = %f, want 0", pct)
	}
}

func TestNewMetricsCollector(t *testing.T) {
	c := NewMetricsCollector(1 * time.Second)
	if c == nil {
		t.Fatal("NewMetricsCollector returned nil")
	}
	if c.sampleInterval != 1*time.Second {
		t.Errorf("sampleInterval = %v, want 1s", c.sampleInterval)
	}
}

func TestMetricsCollector_TaskMethods(t *testing.T) {
	c := NewMetricsCollector(5 * time.Second)

	// 测试任务提交
	c.RecordTaskSubmitted()
	if c.taskSubmitted != 1 {
		t.Errorf("taskSubmitted = %d, want 1", c.taskSubmitted)
	}
	if c.taskInProgress != 1 {
		t.Errorf("taskInProgress = %d, want 1", c.taskInProgress)
	}

	// 测试任务完成
	c.RecordTaskCompleted(100 * time.Millisecond)
	if c.taskCompleted != 1 {
		t.Errorf("taskCompleted = %d, want 1", c.taskCompleted)
	}
	if c.taskInProgress != 0 {
		t.Errorf("taskInProgress = %d, want 0", c.taskInProgress)
	}

	// 测试任务失败
	c.RecordTaskSubmitted()
	c.RecordTaskFailed()
	if c.taskFailed != 1 {
		t.Errorf("taskFailed = %d, want 1", c.taskFailed)
	}
}

func TestMetricsCollector_DownloadMethods(t *testing.T) {
	c := NewMetricsCollector(5 * time.Second)

	// 测试下载开始
	c.RecordDownloadStart()
	if c.downloadActive != 1 {
		t.Errorf("downloadActive = %d, want 1", c.downloadActive)
	}

	// 测试下载完成
	c.RecordDownloadComplete(1024*1024, 2*time.Second)
	if c.downloadActive != 0 {
		t.Errorf("downloadActive = %d, want 0", c.downloadActive)
	}
	if c.downloadTotal != 1024*1024 {
		t.Errorf("downloadTotal = %d, want 1048576", c.downloadTotal)
	}

	// 测试下载错误
	c.RecordDownloadStart()
	c.RecordDownloadError()
	if c.downloadErrors != 1 {
		t.Errorf("downloadErrors = %d, want 1", c.downloadErrors)
	}
}

func TestMetricsCollector_SearchMethods(t *testing.T) {
	c := NewMetricsCollector(5 * time.Second)

	// 测试搜索开始
	c.RecordSearchStart()
	if c.searchActive != 1 {
		t.Errorf("searchActive = %d, want 1", c.searchActive)
	}

	// 测试缓存命中
	c.RecordSearchComplete(true)
	if c.searchActive != 0 {
		t.Errorf("searchActive = %d, want 0", c.searchActive)
	}
	if c.searchCacheHits != 1 {
		t.Errorf("searchCacheHits = %d, want 1", c.searchCacheHits)
	}

	// 测试缓存未命中
	c.RecordSearchStart()
	c.RecordSearchComplete(false)
	if c.searchCacheMisses != 1 {
		t.Errorf("searchCacheMisses = %d, want 1", c.searchCacheMisses)
	}
}

func TestMetricsCollector_Collect(t *testing.T) {
	c := NewMetricsCollector(5 * time.Second)

	// 记录一些数据
	c.RecordTaskSubmitted()
	c.RecordTaskCompleted(100 * time.Millisecond)
	c.RecordDownloadStart()
	c.RecordDownloadComplete(1024*1024, 2*time.Second)
	c.RecordSearchStart()
	c.RecordSearchComplete(true)
	c.SetJSPoolSize(10)
	c.SetJSPoolActive(5)

	// 收集指标
	m := c.Collect()

	// 验证基本指标
	if m.TotalTasks != 1 {
		t.Errorf("TotalTasks = %d, want 1", m.TotalTasks)
	}
	if m.CompletedTasks != 1 {
		t.Errorf("CompletedTasks = %d, want 1", m.CompletedTasks)
	}
	if m.TotalDownloaded != 1024*1024 {
		t.Errorf("TotalDownloaded = %d, want 1048576", m.TotalDownloaded)
	}
	if m.JSPoolSize != 10 {
		t.Errorf("JSPoolSize = %d, want 10", m.JSPoolSize)
	}
	if m.JSPoolActive != 5 {
		t.Errorf("JSPoolActive = %d, want 5", m.JSPoolActive)
	}
	if m.JSPoolUtilization != 50.0 {
		t.Errorf("JSPoolUtilization = %f, want 50.0", m.JSPoolUtilization)
	}
}

func TestGlobalCollector(t *testing.T) {
	// 测试全局收集器存在
	if GlobalCollector == nil {
		t.Error("GlobalCollector should not be nil")
	}

	// 测试便捷函数不 panic
	m := GetMetrics()
	_ = m

	h := GetMetricsHistory()
	if h == nil {
		t.Error("GetMetricsHistory should not return nil")
	}
}
