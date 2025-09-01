package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/LootNex/OrderService/Consumer/internal/errs"
	"github.com/LootNex/OrderService/Consumer/internal/models"
	"github.com/go-redis/redis/v8"
)

type CacheStorage struct {
	rediscache *redis.Client
}

type CacheManager interface {
	SaveOrderCache(ctx context.Context, order models.Order) error
	GetOrderByID(ctx context.Context, orderID string) (models.Order, error)
}

func NewCacheStorage(redisConn *redis.Client) *CacheStorage {
	return &CacheStorage{
		rediscache: redisConn,
	}
}

func (cs *CacheStorage) SaveOrderCache(ctx context.Context, order models.Order) error {

	orderJson, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("cannot marshall order err:%w", err)
	}

	err = cs.rediscache.Set(ctx, order.OrderUID, orderJson, 10*time.Second).Err()
	if err != nil {
		return fmt.Errorf("cannot save order in redis set err:%w", err)
	}

	return nil
}

func (cs *CacheStorage) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {

	order := models.Order{}

	orderJson, err := cs.rediscache.Get(ctx, orderID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return order, errs.ErrOrderNotFound
		} else {
			return order, fmt.Errorf("redis get err:%w", err)
		}
	}

	if err = json.Unmarshal([]byte(orderJson), &order); err != nil {
		return order, fmt.Errorf("cannot unmarshal order err:%w", err)
	}

	return order, nil

}
