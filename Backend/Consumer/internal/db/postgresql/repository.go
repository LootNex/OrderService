package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/LootNex/OrderService/Consumer/internal/models"
	"go.uber.org/zap"
)

type PGStorage struct {
	db  *sql.DB
	log *zap.Logger
}

type RepManager interface {
	SaveNewOrder(ctx context.Context, order models.Order) error
	GetOrderByID(ctx context.Context, orderID string) (models.Order, error)
	GetAllOrderID(ctx context.Context) ([]string, error)
}

func NewPGStorage(db *sql.DB, logg *zap.Logger) *PGStorage {
	return &PGStorage{
		db:  db,
		log: logg,
	}
}

func (pg *PGStorage) SaveNewOrder(ctx context.Context, order models.Order) error {

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction err:%w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				pg.log.Error("rollback failed: %v", zap.Error(rbErr))
			}
		}
	}()

	_, err = tx.ExecContext(ctx, "INSERT INTO Orders VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
		order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)

	if err != nil {
		return fmt.Errorf("cannot insert into table Orders err: %w", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO Delivery(name, phone, zip, city, address, region, email, order_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email, order.OrderUID)

	if err != nil {
		return fmt.Errorf("cannot insert into table Delivery err: %w", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO Payments VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee, order.OrderUID)

	if err != nil {
		return fmt.Errorf("cannot insert into table Payments err: %w", err)
	}

	for _, item := range order.Items {

		_, err = tx.ExecContext(ctx, "INSERT INTO Items VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
			item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status, order.OrderUID)

		if err != nil {
			return fmt.Errorf("cannot insert into table Items err: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("cannot commit transaction err:%w", err)
	}

	return nil

}

func (pg *PGStorage) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {

	var order models.Order
	var items []models.Item

	err := pg.db.QueryRowContext(ctx, "SELECT track_number, entry, locale, internal_signature, customer_id, delivery_service,"+
		"shardkey, sm_id, date_created, oof_shard FROM Orders WHERE order_uid = $1", orderID).Scan(
		&order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID,
		&order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard)

	if err != nil {
		return models.Order{}, fmt.Errorf("cannot scan info from Orders err:%w", err)
	}

	err = pg.db.QueryRowContext(ctx, "SELECT delivery_id, name, phone, zip, city, address, region, email"+
		" FROM Delivery WHERE order_id = $1", orderID).Scan(
		&order.Delivery.Delivery_ID, &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)

	if err != nil {
		return models.Order{}, fmt.Errorf("cannot scan info from Delivery err:%w", err)
	}

	err = pg.db.QueryRowContext(ctx, "SELECT transaction, request_id, currency, provider, amount, payment_dt, bank,"+
		" delivery_cost, goods_total, custom_fee FROM Payments WHERE order_id = $1", orderID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
		&order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee)

	if err != nil {
		return models.Order{}, fmt.Errorf("cannot scan info from Payments err:%w", err)
	}

	rows, err := pg.db.QueryContext(ctx, "SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status"+
		" FROM Items WHERE order_id = $1", orderID)
	if err != nil {
		return models.Order{}, fmt.Errorf("cannot get info about items err:%w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var item models.Item

		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
			return models.Order{}, fmt.Errorf("cannot scan info item err:%w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return models.Order{}, fmt.Errorf("error while scanning rows err:%w", err)
	}

	order.Items = items

	return order, nil

}

func (pg *PGStorage) GetAllOrderID(ctx context.Context) ([]string, error) {

	var IDs []string

	rows, err := pg.db.QueryContext(ctx, "SELECT order_uid FROM Orders")
	if err != nil {
		return nil, fmt.Errorf("cannot get all orderID from Orders err:%w", err)
	}

	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("cannot scan id")
		}

		IDs = append(IDs, id)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while scanning rows err:%w", err)
	}

	return IDs, nil

}
