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
	producer      sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	producerAsync bool
	producerTopic string
)

func InitProducer() {
	kc := conf.GetConfiguration().Kafka
	c := kc.ProducerMessage
	producerTopic = c.Topic

	cfg := newSaramaConfig(kc.Version)
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	cfg.Producer.RequiredAcks = toRequiredAcks(c.RequiredBacks)
	cfg.Producer.Partitioner = sarama.NewHashPartitioner // 按 key 哈希分区
	cfg.Producer.Timeout = 3 * time.Second
	cfg.Producer.Retry.Max = 3

	ctx := context.Background()
	if c.Async {
		producerAsync = true
		ap, err := sarama.NewAsyncProducer(c.Brokers, cfg)
		if err != nil {
			logs.Fatal(ctx, fmt.Sprintf("failed to create async kafka producer: %v", err))
		}
		asyncProducer = ap
		go drainAsyncProducer(ap)
	} else {
		sp, err := sarama.NewSyncProducer(c.Brokers, cfg)
		if err != nil {
			logs.Fatal(ctx, fmt.Sprintf("failed to create sync kafka producer: %v", err))
		}
		producer = sp
	}

	logs.Info(ctx, "Kafka producer initialized successfully")
}

// SendMessage 发送一条消息。key 用于分区路由（按 key 哈希）。
// 同步模式：等待 broker 确认后返回；异步模式：投递到内部队列后立即返回，
// 发送失败会在后台 goroutine 中记录日志。
func SendMessage(ctx context.Context, key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: producerTopic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	if producerAsync {
		select {
		case asyncProducer.Input() <- msg:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	_, _, err := producer.SendMessage(msg)
	return err
}

func GetProducer() sarama.SyncProducer {
	return producer
}

func GetAsyncProducer() sarama.AsyncProducer {
	return asyncProducer
}

func CloseProducer() error {
	if producerAsync && asyncProducer != nil {
		return asyncProducer.Close()
	}
	if producer != nil {
		return producer.Close()
	}
	return nil
}

func toRequiredAcks(n int) sarama.RequiredAcks {
	switch n {
	case -1:
		return sarama.WaitForAll
	case 0:
		return sarama.NoResponse
	default:
		return sarama.WaitForLocal
	}
}

// newSaramaConfig 构造一份带版本号的 sarama 配置，供生产者/消费者复用。
func newSaramaConfig(version string) *sarama.Config {
	cfg := sarama.NewConfig()
	if v, err := sarama.ParseKafkaVersion(version); err != nil {
		logs.Error(context.Background(), fmt.Sprintf("parse kafka version %q failed, use default: %v", version, err))
	} else {
		cfg.Version = v
	}
	return cfg
}

// drainAsyncProducer 消费异步生产者的 Errors/Successes 通道，避免内部队列阻塞。
func drainAsyncProducer(p sarama.AsyncProducer) {
	ctx := context.Background()
	for {
		select {
		case err, ok := <-p.Errors():
			if !ok {
				return
			}
			logs.Error(ctx, fmt.Sprintf("kafka async producer error: %v", err))
		case _, ok := <-p.Successes():
			if !ok {
				return
			}
		}
	}
}
