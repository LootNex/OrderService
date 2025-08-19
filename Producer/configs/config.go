package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Kafka struct {
		Brokers []string
		Topic   string
	}
}

func InitConfig() (*Config, error) {

	viper.SetConfigFile("/ProducerRoot/configs/config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error read in config err: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config err: %v", err)
	}

	return &cfg, nil

}
