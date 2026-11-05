package chat

import (
	"testing"
	"time"

	"tongcheng/library/snowflake"
)

func TestWriteTimeout(t *testing.T) {
	InitMediator()
	snowflake.InitSnowflake(0)

	t.Run("write to channel with timeout", func(t *testing.T) {
		InitMediator()
		// 创建一个带缓冲的通道用于测试
		testChan := make(chan string, 1)

		// 发送一个值到通道，应该成功
		WriteTimeout(testChan, "test message", time.Second*5)

		// 检查消息是否发送成功
		select {
		case msg := <-testChan:
			if msg != "test message" {
				t.Errorf("Expected 'test message', got '%s'", msg)
			}
		default:
			t.Error("Expected message was not sent to channel")
		}
	})

	t.Run("write to full channel with timeout", func(t *testing.T) {
		// 创建一个无缓冲且不被读取的通道，模拟满的通道情况
		fullChan := make(chan string)

		// 记录开始时间
		start := time.Now()

		// 尝试发送值到无法接收的通道，应超时返回
		WriteTimeout(fullChan, "test message", WriteChannelTime)

		// 确保等待了足够的超时时间
		elapsed := time.Since(start)
		if elapsed < WriteChannelTime {
			t.Error("Function returned before timeout")
		}

		// 确保没有阻塞太久
		if elapsed > WriteChannelTime+1*time.Second+time.Millisecond*10 {
			t.Error("Function blocked longer than expected timeout")
		}
	})
}
