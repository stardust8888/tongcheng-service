package kafka

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

// TestSaramaProduceConsumeMultiPartition 用 github.com/IBM/sarama 接入本地 Docker
// Kafka：创建一个多分区 topic，向每个分区写入数据，再为每个分区各起一个消费者
// 消费，最后校验每条消息的内容与分区内顺序。
//
// 运行前提：docker 里的 kafka 已启动并映射到宿主机 9092 端口。
// 可用 KAFKA_TEST_BROKER 环境变量覆盖 broker 地址。
// 未连上 kafka 时测试会自动 Skip，不会误报失败。
func TestSaramaProduceConsumeMultiPartition(t *testing.T) {
	broker := os.Getenv("KAFKA_TEST_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	brokers := []string{broker}

	const (
		numPartitions    = 3
		msgsPerPartition = 5
	)
	total := numPartitions * msgsPerPartition
	topicName := fmt.Sprintf("sarama-test-%d", time.Now().UnixNano())

	/* ---------- 1. sarama 配置 ---------- */
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_0_0_0
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Return.Successes = true
	cfg.Producer.Partitioner = sarama.NewManualPartitioner // 手动指定分区
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	/* ---------- 2. 建 3 分区 topic（探活 + 创建）---------- */
	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	if err != nil {
		t.Skipf("kafka 不可用（NewClusterAdmin 失败），跳过测试: %v", err)
	}
	defer admin.Close()
	defer func() { _ = admin.DeleteTopic(topicName) }() // 清理测试 topic

	if err := admin.CreateTopic(topicName, &sarama.TopicDetail{
		NumPartitions:     int32(numPartitions),
		ReplicationFactor: 1,
	}, false); err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
		t.Fatalf("CreateTopic: %v", err)
	}
	time.Sleep(2 * time.Second) // 等 metadata 传播

	/* ---------- 3. 写入数据：每个分区各写 5 条 ---------- */
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		t.Fatalf("NewSyncProducer: %v", err)
	}
	defer producer.Close()

	sent := make([][]string, numPartitions)
	for p := range numPartitions {
		for i := range msgsPerPartition {
			val := fmt.Sprintf("p%d-m%d", p, i)
			_, _, err := producer.SendMessage(&sarama.ProducerMessage{
				Topic:     topicName,
				Partition: int32(p), // ManualPartitioner 下生效
				Key:       sarama.StringEncoder(fmt.Sprintf("key-%d", p)),
				Value:     sarama.StringEncoder(val),
			})
			if err != nil {
				t.Fatalf("SendMessage p=%d i=%d: %v", p, i, err)
			}
			sent[p] = append(sent[p], val)
		}
	}
	t.Logf("已写入 %d 条消息到 %d 个分区", total, numPartitions)

	/* ---------- 4. 消费：每个分区一个 PartitionConsumer，从最早位点读 ---------- */
	consumer, err := sarama.NewConsumer(brokers, cfg)
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer consumer.Close()

	type result struct {
		partition int
		value     string
	}
	results := make(chan result, total)
	defer close(results)
	for p := range numPartitions {
		go func(partition int) {
			pc, err := consumer.ConsumePartition(topicName, int32(partition), sarama.OffsetOldest)
			if err != nil {
				t.Errorf("ConsumePartition(%d): %v", partition, err)
				return
			}
			defer pc.Close()

			for i := range msgsPerPartition {
				select {
				case msg, ok := <-pc.Messages():
					if !ok {
						t.Errorf("分区 %d 的消息通道提前关闭", partition)
						return
					}
					t.Logf("消费到消息 第%d分区，消息内容%s", partition, string(msg.Value))
					results <- result{partition: partition, value: string(msg.Value)}
				case <-time.After(15 * time.Second):
					t.Errorf("分区 %d 消费超时（已收 %d/%d）", partition, i, msgsPerPartition)
					return
				}
			}
		}(p)
	}

	time.Sleep(10 * time.Second)
}
