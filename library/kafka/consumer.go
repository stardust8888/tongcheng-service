package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"

	"tongcheng/conf"
	"tongcheng/logs"
)

var (
	consumerGroup sarama.ConsumerGroup
	consumerCfg   conf.Consumer
)

func InitConsumer() error {
	kc := conf.GetConfiguration().Kafka
	consumerCfg = kc.ConsumerMessagePush

	cfg := newSaramaConfig(kc.Version)
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = toOffset(consumerCfg.StartOffset)
	if consumerCfg.CommitInterval > 0 {
		cfg.Consumer.Offsets.AutoCommit.Enable = true
		cfg.Consumer.Offsets.AutoCommit.Interval = time.Duration(consumerCfg.CommitInterval) * time.Millisecond
	} else {
		cfg.Consumer.Offsets.AutoCommit.Enable = false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	consumerGroup, err = sarama.NewConsumerGroup(consumerCfg.Brokers, consumerCfg.GroupID, cfg)
	if err != nil {
		return fmt.Errorf("failed to init kafka consumer group: %v", err)
	}

	go func() {
		for e := range consumerGroup.Errors() {
			logs.Error(context.Background(), fmt.Sprintf("kafka consumer group error: %v", e))
		}
	}()

	logs.Info(ctx, "Kafka consumer initialized successfully")
	return nil
}

// Consume 启动消费循环，handler 处理每条消息。
//   - handler 返回 nil 视为处理成功，框架会标记位点；
//   - handler 返回 error 会终止当前批次，从上次提交位点重新消费该消息（at-least-once）。
//
// 提交策略：CommitInterval>0 时按间隔自动提交；=0 时在 rebalance/session 结束时提交。
// 该方法阻塞直到 ctx 结束。
func Consume(ctx context.Context, handler func(*sarama.ConsumerMessage) error) error {
	if consumerGroup == nil {
		return fmt.Errorf("kafka consumer not initialized")
	}
	for {
		h := &consumerHandler{handler: handler}
		if err := consumerGroup.Consume(ctx, []string{consumerCfg.Topic}, h); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			logs.Error(ctx, fmt.Sprintf("kafka consume error, retrying: %v", err))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
			continue
		}
		if h.err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			logs.Error(ctx, fmt.Sprintf("kafka message handler error, retrying from last commit: %v", h.err))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
			continue
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func GetConsumer() sarama.ConsumerGroup {
	return consumerGroup
}

func CloseConsumer() error {
	if consumerGroup != nil {
		return consumerGroup.Close()
	}
	return nil
}

func toOffset(n int64) int64 {
	switch n {
	case -2:
		return sarama.OffsetOldest
	case -1:
		return sarama.OffsetNewest
	default:
		return n
	}
}

type consumerHandler struct {
	handler func(*sarama.ConsumerMessage) error
	err     error
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.handler(msg); err != nil {
			h.err = err
			return err // 退出，下次从上次提交位点重消费该消息
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
