package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LootNex/OrderService/Consumer/internal/errs"
	"github.com/LootNex/OrderService/Consumer/internal/models"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type MockServiceManager struct {
	SaveNewOrderFunc func(ctx context.Context, val models.Validator) error
	GetOrderByIDFunc func(ctx context.Context, orderID string) (models.Order, error)
	LoadCacheFunc    func(ctx context.Context) error
}

func (mSM MockServiceManager) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {
	return mSM.GetOrderByIDFunc(ctx, orderID)
}

func (mSM MockServiceManager) SaveNewOrder(ctx context.Context, val models.Validator) error {
	return mSM.SaveNewOrderFunc(ctx, val)
}

func (mSM MockServiceManager) LoadCache(ctx context.Context) error {
	return mSM.LoadCacheFunc(ctx)
}

func TestGetOrder_Success(t *testing.T) {

	mock := MockServiceManager{
		GetOrderByIDFunc: func(ctx context.Context, orderID string) (models.Order, error) {
			return models.Order{}, nil
		},
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("cannot init logger err: %v", err)
		return
	}

	h := NewHandler(mock, log)

	r := httptest.NewRequest(http.MethodGet, "/order/12345", nil)
	r = mux.SetURLVars(r, map[string]string{"id": "12345"})
	w := httptest.NewRecorder()

	h.GetOrder(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected status %v, expected %v", res.StatusCode, http.StatusOK)
	}

}

func TestGetOrder_NotFound(t *testing.T) {
	mock := MockServiceManager{
		GetOrderByIDFunc: func(ctx context.Context, orderID string) (models.Order, error) {
			return models.Order{}, errs.ErrOrderNotFound
		},
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("cannot init logger err: %v", err)
		return
	}

	h := NewHandler(mock, log)

	r := httptest.NewRequest(http.MethodGet, "/order/12345", nil)
	r = mux.SetURLVars(r, map[string]string{"id": "12345"})
	w := httptest.NewRecorder()

	h.GetOrder(w, r)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("unexpected status %v, expected %v", res.StatusCode, http.StatusOK)
	}

}
