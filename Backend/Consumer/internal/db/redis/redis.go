package redis

import (
	"context"
	"fmt"

	"github.com/LootNex/OrderService/Consumer/configs"
	"github.com/go-redis/redis/v8"
)

func InitRedis(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		// Username:     cfg.Redis.User,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.Timeout,
		WriteTimeout: cfg.Redis.Timeout,
	})

	if err := db.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}

	return db, nil
}
