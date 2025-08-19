package cache

import "github.com/LootNex/OrderService/Consumer/internal/models"

type CacheStorage struct {
	storage map[string]models.Order
}

type CacheManager interface {
	SaveOrderCache(order models.Order)
	GetOrderByID(orderID string) models.Order
}

func NewCacheStorage() *CacheStorage {
	return &CacheStorage{
		storage: make(map[string]models.Order),
	}
}

func (cs *CacheStorage) SaveOrderCache(order models.Order) {

	cs.storage[order.OrderUID] = order

}

func (cs *CacheStorage) GetOrderByID(orderID string) models.Order {

	if order, ok := cs.storage[orderID]; ok {
		return order
	}

	return models.Order{}

}
