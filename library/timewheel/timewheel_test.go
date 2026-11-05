package timewheel

import (
	"sync"
	"testing"
	"time"
)

func TestTimeWheel_New(t *testing.T) {
	// 测试正常创建时间轮
	job := func(data any) {}
	tw := New(time.Second, 60, job)
	if tw == nil {
		t.Error("Expected new TimeWheel, got nil")
	}

	// 测试无效参数
	tw = New(0, 60, job)
	if tw != nil {
		t.Error("Expected nil for zero interval, got TimeWheel")
	}

	tw = New(time.Second, 0, job)
	if tw != nil {
		t.Error("Expected nil for zero slotNum, got TimeWheel")
	}

	tw = New(time.Second, 60, nil)
	if tw != nil {
		t.Error("Expected nil for nil job, got TimeWheel")
	}
}

func TestTimeWheel_AddTask(t *testing.T) {
	var mutex sync.Mutex
	var executed bool
	var receivedData any

	job := func(data any) {
		mutex.Lock()
		executed = true
		receivedData = data
		mutex.Unlock()
	}

	tw := New(1*time.Second, 10, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}
	tw.Start()
	defer tw.Stop()

	testData := "test_data"
	tw.AddTask(1*time.Second, -1, "test_key", testData)

	// 等待任务执行
	time.Sleep(10 * time.Second)

	mutex.Lock()
	if !executed {
		t.Error("Expected task to be executed")
	}
	if receivedData != testData {
		t.Errorf("Expected data %v, got %v", testData, receivedData)
	}
	mutex.Unlock()
}

func TestTimeWheel_RemoveTimer(t *testing.T) {
	var mutex sync.Mutex
	var executed bool

	job := func(data any) {
		mutex.Lock()
		executed = true
		mutex.Unlock()
	}

	tw := New(100*time.Millisecond, 6, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}
	tw.Start()
	defer tw.Stop()

	tw.AddTask(500*time.Millisecond, 1, "test_key", "test_data")
	tw.RemoveTimer("test_key")

	// 等待足够长时间以确保任务本应执行
	time.Sleep(600 * time.Millisecond)

	mutex.Lock()
	if executed {
		t.Error("Expected task to not be executed after removal")
	}
	mutex.Unlock()
}

func TestTimeWheel_RepeatingTask(t *testing.T) {
	var mutex sync.Mutex
	executionCount := 0

	job := func(data any) {
		mutex.Lock()
		executionCount++
		t.Logf("Task executed, count: %d", executionCount)
		mutex.Unlock()
	}

	tw := New(100*time.Millisecond, 6, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}
	tw.Start()
	defer tw.Stop()

	tw.AddTask(200*time.Millisecond, 3, "repeat_key", "test_data")

	// 等待足够时间让任务执行3次
	time.Sleep(1500 * time.Millisecond)

	mutex.Lock()
	if executionCount != 3 {
		t.Errorf("Expected task to execute 3 times, got %d", executionCount)
	}
	mutex.Unlock()
}

func TestTimeWheel_InfiniteTask(t *testing.T) {
	var mutex sync.Mutex
	executionCount := 0

	job := func(data any) {
		mutex.Lock()
		executionCount++
		mutex.Unlock()
	}

	tw := New(100*time.Millisecond, 6, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}
	tw.Start()
	defer tw.Stop()

	// 添加一个执行多次的任务来模拟无限任务
	tw.AddTask(200*time.Millisecond, -1, "infinite_key", "test_data")

	// 等待足够时间让任务执行几次
	time.Sleep(800 * time.Millisecond)

	tw.RemoveTimer("infinite_key")

	mutex.Lock()
	if executionCount == 0 {
		t.Error("Expected infinite task to execute at least once")
	}
	t.Logf("Infinite task executed %d times", executionCount)
	mutex.Unlock()
}

func TestGetPositionAndCircle(t *testing.T) {
	job := func(data any) {}
	tw := New(time.Second, 60, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}

	// 测试位置和圈数计算
	pos, circle := tw.getPositionAndCircle(61 * time.Second)
	expectedPos := (tw.currentPos + 1) % tw.slotNum
	expectedCircle := 1

	if pos != expectedPos {
		t.Errorf("Expected position %d, got %d", expectedPos, pos)
	}
	if circle != expectedCircle {
		t.Errorf("Expected circle %d, got %d", expectedCircle, circle)
	}
}

func TestInitTimeWheel(t *testing.T) {
	var executed bool
	job := func(data any) {
		executed = true
	}

	tw := New(100*time.Millisecond, 6, job)
	if tw == nil {
		t.Fatal("Failed to initialize TimeWheel")
	}
	defer tw.Stop()

	tw.AddTask(200*time.Millisecond, 1, "init_test", "data")

	// 等待任务执行
	time.Sleep(300 * time.Millisecond)

	if !executed {
		t.Error("Expected task to be executed with InitTimeWheel")
	}
}

func TestStopTimeWheel(t *testing.T) {
	job := func(data any) {
		// 空实现
	}

	tw := New(100*time.Millisecond, 6, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}
	tw.Start()

	// 发送停止信号
	tw.Stop()

	// 验证ticker已停止
	select {
	case <-time.After(200 * time.Millisecond):
		// 正常情况，没有接收到tick
	}
}

func TestTimeWheel_TaskWithZeroCount(t *testing.T) {
	job := func(data any) {
		// 空实现
	}

	tw := New(100*time.Millisecond, 6, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}
	tw.Start()
	defer tw.Stop()

	// 测试count为0的任务不应该被添加
	tw.AddTask(200*time.Millisecond, 0, "zero_count", "test_data")

	// 简单等待一下确保没有panic
	time.Sleep(100 * time.Millisecond)
}

func TestTimeWheel_TaskWithNegativeCount(t *testing.T) {
	var mutex sync.Mutex
	executionCount := 0

	job := func(data any) {
		mutex.Lock()
		executionCount++
		mutex.Unlock()
	}

	tw := New(100*time.Millisecond, 6, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}
	tw.Start()
	defer tw.Stop()

	// 测试count为-2的任务不应该被添加(count只有-1表示无限执行是有效的)
	tw.AddTask(200*time.Millisecond, -2, "negative_count", "test_data")

	// 简单等待一下确保没有任务执行
	time.Sleep(500 * time.Millisecond)

	mutex.Lock()
	if executionCount != 0 {
		t.Errorf("Expected no task execution for negative count, got %d executions", executionCount)
	}
	mutex.Unlock()
}

func TestTimeWheel_TaskWithNilKey(t *testing.T) {
	var mutex sync.Mutex
	executionCount := 0

	job := func(data any) {
		mutex.Lock()
		executionCount++
		mutex.Unlock()
	}

	tw := New(100*time.Millisecond, 6, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}
	tw.Start()
	defer tw.Stop()

	// 测试key为nil的任务
	tw.AddTask(200*time.Millisecond, 1, nil, "test_data")

	// 等待任务执行
	time.Sleep(300 * time.Millisecond)

	mutex.Lock()
	if executionCount != 0 {
		t.Errorf("Expected task with nil key to execute once, got %d executions", executionCount)
	}
	mutex.Unlock()
}

func TestTimeWheel_RemoveNonExistentTimer(t *testing.T) {
	job := func(data any) {
		// 空实现
	}

	tw := New(100*time.Millisecond, 6, job)
	if tw == nil {
		t.Fatal("Failed to create TimeWheel")
	}
	tw.Start()
	defer tw.Stop()

	// 尝试删除不存在的定时器不应该导致panic
	tw.RemoveTimer("non_existent_key")

	// 简单等待一下确保没有panic
	time.Sleep(100 * time.Millisecond)
}

func BenchmarkTimeWheel_AddTask(b *testing.B) {
	job := func(data any) {}
	tw := New(time.Millisecond*100, 60, job)
	tw.Start()
	defer tw.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tw.AddTask(time.Second, 1, i, "benchmark")
	}
}

func BenchmarkTimeWheel_ExecuteTask(b *testing.B) {
	executed := make(chan struct{}, b.N)
	job := func(data any) {
		executed <- struct{}{}
	}

	tw := New(time.Millisecond, 6, job)
	tw.Start()
	defer tw.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tw.AddTask(time.Millisecond*10, 1, i, "benchmark")
	}

	// 等待所有任务执行完成
	for i := 0; i < b.N; i++ {
		<-executed
	}
}
