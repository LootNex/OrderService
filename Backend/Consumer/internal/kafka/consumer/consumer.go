package consumer

import (
	"context"
	"encoding/json"

	"github.com/LootNex/OrderService/Consumer/internal/models"
	"github.com/LootNex/OrderService/Consumer/internal/service"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func StartConsumer(ctx context.Context, topic string, brokers []string, serv service.ServiceManager, log *zap.Logger) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     "order-service",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3,
		MaxBytes:    10e6,
	})
	defer func() {
		if err := r.Close(); err != nil {
			log.Error("failed to close Kafka reader", zap.Error(err))
		}
	}()

	log.Info("Kafka consumer started")

	for {
		select {
		case <-ctx.Done():
			log.Info("Kafka consumer stopping gracefully...")
			return
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				log.Warn("Ошибка чтения из Kafka", zap.Error(err))
				continue
			}

			var order models.Order

			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Warn("Ошибка парсинга JSON", zap.Error(err))
				continue
			}

			if err := serv.SaveNewOrder(ctx, &order); err != nil {
				log.Warn("Ошибка сохранения заказа", zap.Error(err))

			} else {
				if err = r.CommitMessages(ctx, msg); err != nil {
					log.Error("error commit message err", zap.Error(err))
				}

				log.Info("GET ORDER", zap.String("order_uid", order.OrderUID))

			}
		}
	}
}
