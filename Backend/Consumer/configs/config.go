package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port string
	}
	Postgres struct {
		Host     string
		Port     int
		User     string
		Password string
		DBname   string
	}
	Kafka struct {
		Brokers []string
		Topic   string
	}
}

func InitConfig() (*Config, error) {

	viper.SetConfigFile("/Consumer/configs/config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read in config err: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config err: %v", err)
	}

	return &cfg, nil

}
