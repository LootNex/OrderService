package service

import (
	"context"

	"github.com/LootNex/OrderService/Consumer/internal/cache"
	"github.com/LootNex/OrderService/Consumer/internal/models"
	"github.com/LootNex/OrderService/Consumer/internal/repository"
)

type OrderService struct {
	Rep  repository.RepManager
	Cach cache.CacheManager
}

type ServiceManager interface {
	SaveNewOrder(ctx context.Context, order models.Order) error
	GetOrderByID(ctx context.Context, orderID string) (models.Order, error)
	LoadCache(ctx context.Context) error
}

func NewOrderService(rep repository.RepManager, cach cache.CacheManager) *OrderService {
	return &OrderService{
		Rep:  rep,
		Cach: cach,
	}
}

func (os *OrderService) SaveNewOrder(ctx context.Context, order models.Order) error {

	err := os.Rep.SaveNewOrder(ctx, order)
	if err != nil {
		return err
	}

	os.Cach.SaveOrderCache(order)

	return nil

}

func (os *OrderService) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {

	orderData := os.Cach.GetOrderByID(orderID)

	if orderData.OrderUID != "" {
		return orderData, nil
	}

	orderData, err := os.Rep.GetOrderByID(ctx, orderID)
	if err != nil {
		return models.Order{}, err
	}

	os.Cach.SaveOrderCache(orderData)

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

		os.Cach.SaveOrderCache(orderData)
	}

	return nil

}
