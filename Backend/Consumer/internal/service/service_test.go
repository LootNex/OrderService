package service

import (
	"context"
	"errors"
	"testing"

	"github.com/LootNex/OrderService/Consumer/internal/errs"
	"github.com/LootNex/OrderService/Consumer/internal/models"
	"go.uber.org/zap"
)

type MockRepManager struct {
	SaveNewOrderFunc  func(ctx context.Context, order models.Order) error
	GetOrderByIDFunc  func(ctx context.Context, orderID string) (models.Order, error)
	GetAllOrderIDFunc func(ctx context.Context) ([]string, error)
}

type MockCacheManager struct {
	SaveOrderCacheFunc func(ctx context.Context, order models.Order) error
	GetOrderByIDFunc   func(ctx context.Context, orderID string) (models.Order, error)
}

type MockValidator struct {
	ValidateFunc func() error
}

func (mRP MockRepManager) SaveNewOrder(ctx context.Context, order models.Order) error {
	return mRP.SaveNewOrderFunc(ctx, order)
}

func (mRP MockRepManager) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {
	return mRP.GetOrderByIDFunc(ctx, orderID)
}

func (mRP MockRepManager) GetAllOrderID(ctx context.Context) ([]string, error) {
	return mRP.GetAllOrderIDFunc(ctx)
}

func (mCM MockCacheManager) SaveOrderCache(ctx context.Context, order models.Order) error {
	return mCM.SaveOrderCacheFunc(ctx, order)
}

func (mCM MockCacheManager) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {
	return mCM.GetOrderByIDFunc(ctx, orderID)
}

func (MV MockValidator) Validate() error {
	return MV.ValidateFunc()
}

func TestSaveNewOrder(t *testing.T) {

	tests := []struct {
		name      string
		validator models.Validator
		repErr    error
		cacheErr  error
		wantErr   bool
	}{
		{
			name:      "success",
			validator: MockValidator{ValidateFunc: func() error { return nil }},
			repErr:    nil,
			cacheErr:  nil,
			wantErr:   false,
		},
		{
			name:      "invalid validator",
			validator: MockValidator{ValidateFunc: func() error { return errors.New("order_uid is required") }},
			repErr:    nil,
			cacheErr:  nil,
			wantErr:   true,
		},
		{
			name: "invalid RepManager",
			validator: &models.Order{
				OrderUID:        "123",
				TrackNumber:     "TRACK123",
				CustomerID:      "cust1",
				DeliveryService: "meest",
				DateCreated:     "2021-11-26T06:22:19Z",
				Delivery: models.Delivery{
					Name: "Test", Phone: "123", Email: "test@test.com",
				},
				Payment: models.Payment{
					Transaction: "tr1", Amount: 100, Currency: "USD",
				},
				Items: []models.Item{
					{ChrtID: 1, TrackNumber: "T1", Price: 100, TotalPrice: 100},
				},
			},
			repErr:   errors.New("cannot save in DB"),
			cacheErr: nil,
			wantErr:  true,
		},
		{
			name: "invalid CacheManager",
			validator: &models.Order{
				OrderUID:        "123",
				TrackNumber:     "TRACK123",
				CustomerID:      "cust1",
				DeliveryService: "meest",
				DateCreated:     "2021-11-26T06:22:19Z",
				Delivery: models.Delivery{
					Name: "Test", Phone: "123", Email: "test@test.com",
				},
				Payment: models.Payment{
					Transaction: "tr1", Amount: 100, Currency: "USD",
				},
				Items: []models.Item{
					{ChrtID: 1, TrackNumber: "T1", Price: 100, TotalPrice: 100},
				},
			},
			repErr:   nil,
			cacheErr: errors.New("cannot save in redis"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			log, err := zap.NewDevelopment()
			if err != nil {
				t.Errorf("cannot init logger err: %v", err)
			}
			orderServ := OrderService{
				Rep:  MockRepManager{SaveNewOrderFunc: func(ctx context.Context, order models.Order) error { return tt.repErr }},
				Cach: MockCacheManager{SaveOrderCacheFunc: func(ctx context.Context, order models.Order) error { return tt.cacheErr }},
				log:  log,
			}

			err = orderServ.SaveNewOrder(context.Background(), tt.validator)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected: %v, got: %v", tt.wantErr, err)
			}
		})
	}

}

func TestGetOrderByID(t *testing.T) {
	tests := []struct {
		name              string
		orderID           string
		cacheGetOrderErr  error
		cacheSaveOrderErr error
		repGetOrderErr    error
		wantErr           bool
	}{
		{
			name:              "success",
			orderID:           "123",
			cacheGetOrderErr:  nil,
			cacheSaveOrderErr: nil,
			repGetOrderErr:    nil,
			wantErr:           false,
		},
		{
			name:              "invalid cacheGetOrder",
			orderID:           "123",
			cacheGetOrderErr:  errs.ErrOrderNotFound,
			cacheSaveOrderErr: nil,
			repGetOrderErr:    nil,
			wantErr:           false,
		},
		{
			name:              "invald cacheSaveOrder",
			orderID:           "123",
			cacheGetOrderErr:  nil,
			cacheSaveOrderErr: errors.New("cannot save"),
			repGetOrderErr:    nil,
			wantErr:           false,
		},
		{
			name:              "invalid repGetOrder",
			orderID:           "123",
			cacheGetOrderErr:  errors.New("no sucn order"),
			cacheSaveOrderErr: nil,
			repGetOrderErr:    errors.New("cannot save"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			log, err := zap.NewDevelopment()
			if err != nil {
				t.Errorf("cannot init logger err: %v", err)
			}
			orderServ := OrderService{
				Rep: MockRepManager{GetOrderByIDFunc: func(ctx context.Context, orderID string) (models.Order, error) {
					return models.Order{}, tt.repGetOrderErr
				}},
				Cach: MockCacheManager{
					SaveOrderCacheFunc: func(ctx context.Context, order models.Order) error { return tt.cacheSaveOrderErr },
					GetOrderByIDFunc: func(ctx context.Context, orderID string) (models.Order, error) {
						return models.Order{}, tt.cacheGetOrderErr
					}},
				log: log,
			}
			_, err = orderServ.GetOrderByID(context.Background(), tt.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestLoadCache(t *testing.T) {
	tests := []struct {
		name              string
		orders            []string
		repGetOrderErr    error
		repGetAllOrderErr error
		wantErr           bool
	}{
		{
			name:              "success",
			orders:            []string{"123", "124", "125"},
			repGetAllOrderErr: nil,
			repGetOrderErr:    nil,
			wantErr:           false,
		},
		{
			name:              "invalid repGetAllOrders",
			orders:            []string{"123", "124", "125"},
			repGetAllOrderErr: errors.New("cannot get orders"),
			repGetOrderErr:    nil,
			wantErr:           true,
		},
		{
			name:              "invalid repGetOrder",
			orders:            []string{"123", "124", "125"},
			repGetAllOrderErr: nil,
			repGetOrderErr:    errors.New("cannot get order"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			log, err := zap.NewDevelopment()
			if err != nil {
				t.Errorf("cannot init logger err: %v", err)
			}
			orderServ := OrderService{
				Rep: MockRepManager{
					GetOrderByIDFunc: func(ctx context.Context, orderID string) (models.Order, error) {
						return models.Order{}, tt.repGetOrderErr
					},
					GetAllOrderIDFunc: func(ctx context.Context) ([]string, error) {
						return tt.orders, tt.repGetAllOrderErr
					}},
				Cach: MockCacheManager{SaveOrderCacheFunc: func(ctx context.Context, order models.Order) error {
					return nil
				}},
				log: log,
			}
			err = orderServ.LoadCache(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("expected: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}
