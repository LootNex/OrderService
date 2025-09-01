package server

import (
	"fmt"
	"time"

	"github.com/LootNex/OrderService/Producer/configs"
	"github.com/LootNex/OrderService/Producer/internal/kafka"
	"github.com/LootNex/OrderService/Producer/internal/models"
	"github.com/brianvoe/gofakeit/v7"
)

func StartServer() error {

	config, err := configs.InitConfig()
	if err != nil {
		return err
	}

	for _, broker := range config.Kafka.Brokers {
		err = kafka.EnsureTopic(broker, config.Kafka.Topic, 1, 1)
		if err != nil {
			fmt.Println("Warning: topic creation:", err)
		}
	}

	producer := kafka.NewKafkaProducer(config.Kafka.Brokers, config.Kafka.Topic)
	defer producer.Writer.Close()

	if err = gofakeit.Seed(time.Now().UnixNano()); err != nil {
		return fmt.Errorf("cannot run gofakeit err:%w", err)
	}

	for i := 1; i <= 10; i++ {
		order := models.Order{
			OrderUID:    gofakeit.UUID(),
			TrackNumber: gofakeit.Regex("[A-Z0-9]{10}"),
			Entry:       "WBIL",
			Delivery: models.Delivery{
				Name:    gofakeit.Name(),
				Phone:   gofakeit.Phone(),
				Zip:     gofakeit.Zip(),
				City:    gofakeit.City(),
				Address: gofakeit.Street(),
				Region:  gofakeit.State(),
				Email:   gofakeit.Email(),
			},
			Payment: models.Payment{
				Transaction:  gofakeit.UUID(),
				Currency:     gofakeit.CurrencyShort(),
				Provider:     "wbpay",
				Amount:       int(gofakeit.Price(100, 5000)),
				PaymentDT:    time.Now().Unix(),
				Bank:         gofakeit.Company(),
				DeliveryCost: gofakeit.Number(100, 2000),
				GoodsTotal:   gofakeit.Number(100, 2000),
				CustomFee:    0,
			},
			Items: []models.Item{
				{
					ChrtID:      gofakeit.Number(1000000, 9999999),
					TrackNumber: gofakeit.LetterN(12),
					Price:       gofakeit.Number(100, 1000),
					RID:         gofakeit.UUID(),
					Name:        gofakeit.ProductName(),
					Sale:        gofakeit.Number(0, 50),
					Size:        gofakeit.RandomString([]string{"S", "M", "L"}),
					TotalPrice:  gofakeit.Number(100, 2000),
					NmID:        gofakeit.Number(100000, 999999),
					Brand:       gofakeit.Company(),
					Status:      202,
				},
			},
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        gofakeit.Username(),
			DeliveryService:   gofakeit.Company(),
			ShardKey:          fmt.Sprintf("%d", gofakeit.Number(1, 10)),
			SmID:              gofakeit.Number(1, 1000),
			DateCreated:       time.Now().Format(time.RFC3339),
			OofShard:          fmt.Sprintf("%d", gofakeit.Number(1, 5)),
		}

		err := producer.Send(order)
		if err != nil {
			return err
		}

		time.Sleep(5 * time.Second)

	}

	return nil

}
