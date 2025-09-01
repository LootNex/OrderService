package redis

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/LootNex/OrderService/Consumer/internal/models"
	"github.com/go-redis/redis/v8"
)

type MockRedisComander struct {
	SetFunc func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	GetFunc func(ctx context.Context, key string) *redis.StringCmd
}

func (mRC MockRedisComander) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return mRC.SetFunc(ctx, key, value, expiration)
}

func (mRC MockRedisComander) Get(ctx context.Context, key string) *redis.StringCmd {
	return mRC.GetFunc(ctx, key)
}

func TestSaveOrderCache(t *testing.T) {

	tests := []struct {
		name    string
		order   models.Order
		setErr  error
		wantErr bool
	}{
		{
			name:    "success",
			order:   models.Order{},
			setErr:  nil,
			wantErr: false,
		},
		{
			name:    "invalid Set",
			order:   models.Order{},
			setErr:  errors.New("cannot set"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CachSt := CacheStorage{
				rediscache: MockRedisComander{
					SetFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
						cmd := redis.NewStatusCmd(ctx)
						if tt.setErr != nil {
							cmd.SetErr(tt.setErr)
						} else {
							cmd.SetVal("OK")
						}
						return cmd
					},
				},
			}
			err := CachSt.SaveOrderCache(context.Background(), tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected err:%v, got err:%v", tt.wantErr, err)
			}
		})
	}

}

func TestGetOrderByID(t *testing.T) {
	tests := []struct {
		name    string
		order   models.Order
		getErr  error
		wantErr bool
	}{
		{
			name:    "success",
			order:   models.Order{OrderUID: "123"},
			getErr:  nil,
			wantErr: false,
		},
		{
			name:    "invalid Get",
			order:   models.Order{OrderUID: "456"},
			getErr:  errors.New("cannot get"),
			wantErr: true,
		},
		{
			name:    "invalid order",
			order:   models.Order{OrderUID: "789"},
			getErr:  redis.Nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CachSt := CacheStorage{
				rediscache: MockRedisComander{
					GetFunc: func(ctx context.Context, key string) *redis.StringCmd {
						cmd := redis.NewStringCmd(ctx)
						if tt.getErr != nil {
							cmd.SetErr(tt.getErr)
						} else {
							orderJSON, _ := json.Marshal(tt.order)
							cmd.SetVal(string(orderJSON))
						}
						return cmd
					},
				},
			}
			_, err := CachSt.GetOrderByID(context.Background(), tt.order.OrderUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected err:%v, got err:%v", tt.wantErr, err)
			}
		})
	}
}
