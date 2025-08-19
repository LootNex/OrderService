package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LootNex/OrderService/Consumer/internal/models"
	"github.com/LootNex/OrderService/Consumer/internal/service"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func StartConsumer(ctx context.Context, topic string, brokers []string, serv service.ServiceManager, logger *zap.Logger) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     "order-service",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3,
		MaxBytes:    10e6,
	})
	defer r.Close()

	logger.Info("Kafka consumer started")

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			logger.Warn(fmt.Sprintf("Ошибка чтения из Kafka: %v", err))
			continue
		}

		var order models.Order

		if err := json.Unmarshal(msg.Value, &order); err != nil {
			logger.Warn(fmt.Sprintf("Ошибка парсинга JSON: %v", err))
			continue
		}

		if err := serv.SaveNewOrder(ctx, order); err != nil {
			logger.Warn(fmt.Sprintf("Ошибка сохранения заказа: %v", err))

		} else {
			logger.Info(fmt.Sprintf("Заказ %s обработан", order.OrderUID))
			r.CommitMessages(ctx, msg)

			logger.Info(fmt.Sprintf("GET ORDER: %s", order.OrderUID))

		}

	}
}
