package service

import (
	"context"
	"errors"

	"github.com/LootNex/OrderService/Consumer/internal/db/postgresql"
	"github.com/LootNex/OrderService/Consumer/internal/db/redis"
	"github.com/LootNex/OrderService/Consumer/internal/errs"
	"github.com/LootNex/OrderService/Consumer/internal/models"
	"go.uber.org/zap"
)

type OrderService struct {
	Rep  postgresql.RepManager
	Cach redis.CacheManager
	log  *zap.Logger
}

type ServiceManager interface {
	SaveNewOrder(ctx context.Context, val models.Validator) error
	GetOrderByID(ctx context.Context, orderID string) (models.Order, error)
	LoadCache(ctx context.Context) error
}

func NewOrderService(rep postgresql.RepManager, cach redis.CacheManager, logg *zap.Logger) *OrderService {
	return &OrderService{
		Rep:  rep,
		Cach: cach,
		log:  logg,
	}
}

func (os *OrderService) SaveNewOrder(ctx context.Context, val models.Validator) error {

	if err := val.Validate(); err != nil {
		return err
	}

	switch order := val.(type) {
	case *models.Order:
		if err := os.Rep.SaveNewOrder(ctx, *order); err != nil {
			return err
		}

		if err := os.Cach.SaveOrderCache(ctx, *order); err != nil {
			return err
		}
	}

	return nil

}

func (os *OrderService) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {

	orderData, err := os.Cach.GetOrderByID(ctx, orderID)
	if err == nil {
		return orderData, nil
	} else if !errors.Is(err, errs.ErrOrderNotFound) {
		return orderData, err
	}

	orderData, err = os.Rep.GetOrderByID(ctx, orderID)
	if err != nil {
		return models.Order{}, err
	}

	if err = os.Cach.SaveOrderCache(ctx, orderData); err != nil {
		os.log.Warn("cannot save order in cache", zap.Error(err))
	}

	return orderData, nil

}

func (os *OrderService) LoadCache(ctx context.Context) error {

	Ids, err := os.Rep.GetAllOrderID(ctx)
	if err != nil {
		return err
	}

	for _, id := range Ids {

		orderData, err := os.Rep.GetOrderByID(ctx, id)
		if err != nil {
			return err
		}

		if err = os.Cach.SaveOrderCache(ctx, orderData); err != nil {
			os.log.Warn("cannot save order in cache", zap.Error(err))
		}
	}

	return nil

}
