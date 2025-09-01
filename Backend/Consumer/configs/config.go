package config

import (
	"fmt"
	"time"

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
	Redis struct {
		Addr     string
		Password string
		// User        string
		DB          int
		MaxRetries  int
		DialTimeout time.Duration
		Timeout     time.Duration
	}
	Kafka struct {
		Brokers []string
		Topic   string
	}
}

func InitConfig() (*Config, error) {

	viper.SetConfigFile("/Consumer/configs/config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read in config err: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config err: %w", err)
	}

	return &cfg, nil

}
