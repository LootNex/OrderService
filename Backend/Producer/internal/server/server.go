package server

import (
	"fmt"
	"time"

	"github.com/LootNex/OrderService/Producer/configs"
	"github.com/LootNex/OrderService/Producer/internal/kafka"
	"github.com/LootNex/OrderService/Producer/internal/models"
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

	time.Sleep(5 * time.Second)

	for i := 1; i <= 10; i++ {
		order := models.Order{
			OrderUID:    fmt.Sprintf("b563feb7b2b84b6test%d", i),
			TrackNumber: fmt.Sprintf("WBILMTESTTRACK%d", i),
			Entry:       "WBIL",
			Delivery: models.Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: models.Payment{
				Transaction:  fmt.Sprintf("b563feb7b2b84b6test%d", i),
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1817,
				PaymentDT:    1637907727,
				Bank:         "alpha",
				DeliveryCost: 1500,
				GoodsTotal:   317,
				CustomFee:    0,
			},
			Items: []models.Item{
				{
					ChrtID:      9934930 + i,
					TrackNumber: fmt.Sprintf("WBILMTESTTRACK%d", i),
					Price:       453,
					RID:         fmt.Sprintf("ab4219087a764ae0btest%d", i),
					Name:        "Mascaras",
					Sale:        30,
					Size:        "0",
					TotalPrice:  317,
					NmID:        2389212,
					Brand:       "Vivienne Sabo",
					Status:      202,
				},
			},
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "test",
			DeliveryService:   "meest",
			ShardKey:          "9",
			SmID:              99,
			DateCreated:       "2021-11-26T06:22:19Z",
			OofShard:          "1",
		}

		err := producer.Send(order)
		if err != nil {
			return err
		}

		time.Sleep(5 * time.Second)

	}

	return nil

}
