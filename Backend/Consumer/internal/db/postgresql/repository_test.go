package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LootNex/OrderService/Consumer/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type MockSqlDB struct {
	BeginTxFunc         func(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	QueryRowContextFunc func(ctx context.Context, query string, args ...any) *sql.Row
	QueryContextFunc    func(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContextFunc     func(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (ms MockSqlDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return ms.BeginTxFunc(ctx, opts)
}

func (ms MockSqlDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return ms.QueryRowContextFunc(ctx, query, args...)
}

func (ms MockSqlDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return ms.QueryContextFunc(ctx, query, args...)
}

func (ms MockSqlDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return ms.ExecContextFunc(ctx, query, args...)
}

func TestSaveNewOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	log := zaptest.NewLogger(t)

	storage := NewPGStorage(db, log)

	order := models.Order{
		OrderUID: "123",
		Delivery: models.Delivery{
			Name: "Test User",
		},
		Payment: models.Payment{
			Transaction: "tx_1",
		},
		Items: []models.Item{
			{ChrtID: 1, Name: "Item 1"},
		},
	}

	mock.ExpectBegin()

	mock.ExpectExec("INSERT INTO Orders").
		WithArgs(order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
			order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO Delivery").
		WithArgs(order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
			order.Delivery.Address, order.Delivery.Region, order.Delivery.Email, order.OrderUID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO Payments").
		WithArgs(order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
			order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
			order.Payment.GoodsTotal, order.Payment.CustomFee, order.OrderUID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO Items").
		WithArgs(order.Items[0].ChrtID, order.Items[0].TrackNumber, order.Items[0].Price, order.Items[0].RID,
			order.Items[0].Name, order.Items[0].Sale, order.Items[0].Size, order.Items[0].TotalPrice,
			order.Items[0].NmID, order.Items[0].Brand, order.Items[0].Status, order.OrderUID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = storage.SaveNewOrder(context.Background(), order)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetOrderByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock db: %v", err)
	}
	defer db.Close()

	log, _ := zap.NewDevelopment()
	pg := NewPGStorage(db, log)

	orderID := "order123"

	mockOrders := sqlmock.NewRows([]string{
		"track_number", "entry", "locale", "internal_signature", "customer_id", "delivery_service",
		"shardkey", "sm_id", "date_created", "oof_shard",
	}).AddRow(
		"TRACK123", "entry1", "en", "sig123", "cust1", "dservice",
		"shard1", 1, "2025-09-01T10:00:00Z", "oof1",
	)
	mock.ExpectQuery("SELECT track_number, entry, locale, internal_signature, customer_id, delivery_service,shardkey, sm_id, date_created, oof_shard FROM Orders").
		WithArgs(orderID).WillReturnRows(mockOrders)

	mockDelivery := sqlmock.NewRows([]string{"delivery_id", "name", "phone", "zip", "city", "address", "region", "email"}).
		AddRow(1, "John Doe", "123456", "11111", "City", "Street 1", "Region", "john@test.com")
	mock.ExpectQuery("SELECT delivery_id, name, phone, zip, city, address, region, email FROM Delivery").
		WithArgs(orderID).WillReturnRows(mockDelivery)

	mockPayments := sqlmock.NewRows([]string{"transaction", "request_id", "currency", "provider", "amount", "payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee"}).
		AddRow("trx123", "req1", "USD", "prov1", 100.0, 1234567890, "bank1", 10.0, 90.0, 0.0)
	mock.ExpectQuery("SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM Payments").
		WithArgs(orderID).WillReturnRows(mockPayments)

	mockItems := sqlmock.NewRows([]string{"chrt_id", "track_number", "price", "rid", "name", "sale", "size", "total_price", "nm_id", "brand", "status"}).
		AddRow(1, "T1", 100.0, "rid1", "ItemName", 0.0, "M", 100.0, 101, "BrandX", 1)
	mock.ExpectQuery("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM Items").
		WithArgs(orderID).WillReturnRows(mockItems)

	ctx := context.Background()
	order, err := pg.GetOrderByID(ctx, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.TrackNumber != "TRACK123" {
		t.Errorf("expected track_number TRACK123, got %v", order.TrackNumber)
	}
	if order.Delivery.Name != "John Doe" {
		t.Errorf("expected delivery name John Doe, got %v", order.Delivery.Name)
	}
	if len(order.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(order.Items))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
func TestGetAllOrderByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock db: %v", err)
	}
	defer db.Close()

	log, _ := zap.NewDevelopment()
	pg := NewPGStorage(db, log)

	orderID := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}

	mockOrderIDs := sqlmock.NewRows([]string{"order_uid"})
	for i := 1; i <= 9; i++ {
		mockOrderIDs.AddRow(fmt.Sprint(i))
	}

	mock.ExpectQuery("SELECT order_uid FROM Orders").WillReturnRows(mockOrderIDs)

	ctx := context.Background()
	orderIds, err := pg.GetAllOrderID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(orderIds, orderID) {
		t.Errorf("expected %v got %v", orderID, orderIds)
	}

}
