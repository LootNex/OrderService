package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LootNex/OrderService/Producer/internal/models"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type KafkaProducer struct {
	Writer *kafka.Writer
}

type KafkaManager interface {
	Send(order models.Order) error
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		Writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:          brokers,
			Topic:            topic,
			Balancer:         &kafka.LeastBytes{},
			RequiredAcks:     int(kafka.RequireAll),
			CompressionCodec: &compress.SnappyCodec,
			BatchSize:        100,
			Async:            false,
		}),
	}
}

func EnsureTopic(broker, topic string, partitions, replication int) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replication,
	})
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}
	return nil
}

func (k KafkaProducer) Send(order models.Order) error {

	msg, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("cannot marshal order %w", err)
	}
	return k.Writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: msg,
		},
	)
}
