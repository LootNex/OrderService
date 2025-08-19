package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LootNex/OrderService/Consumer/internal/models"
)

type PGStorage struct {
	db *sql.DB
}

type RepManager interface {
	SaveNewOrder(ctx context.Context, order models.Order) error
	GetOrderByID(ctx context.Context, orderID string) (models.Order, error)
	GetAllOrderID(ctx context.Context) ([]string, error)
}

func NewPGStorage(db *sql.DB) *PGStorage {
	return &PGStorage{
		db: db,
	}
}

func (pg *PGStorage) SaveNewOrder(ctx context.Context, order models.Order) error {

	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction err:%v", err)
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO Orders VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
		order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)

	if err != nil {
		return fmt.Errorf("cannot insert into table Orders err: %v", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO Delivery(name, phone, zip, city, address, region, email, order_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email, order.OrderUID)

	if err != nil {
		return fmt.Errorf("cannot insert into table Delivery err: %v", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO Payments VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee, order.OrderUID)

	if err != nil {
		return fmt.Errorf("cannot insert into table Payments err: %v", err)
	}

	for _, item := range order.Items {

		_, err = tx.ExecContext(ctx, "INSERT INTO Items VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
			item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status, order.OrderUID)

		if err != nil {
			return fmt.Errorf("cannot insert into table Items err: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("cannot commit transaction err:%v", err)
	}

	return nil

}

func (pg *PGStorage) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {

	var order models.Order
	var items []models.Item

	order.OrderUID = orderID

	err := pg.db.QueryRowContext(ctx, "SELECT * FROM Orders WHERE order_id = $1", orderID).Scan(
		&order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID,
		&order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard)

	if err != nil {
		return models.Order{}, fmt.Errorf("cannot scan info from Orders err:%v", err)
	}

	err = pg.db.QueryRowContext(ctx, "SELECT * FROM Delivery WHERE order_id = $1", orderID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)

	if err != nil {
		return models.Order{}, fmt.Errorf("cannot scan info from Delivery err:%v", err)
	}

	err = pg.db.QueryRowContext(ctx, "SELECT * FROM Payments WHERE order_id = $1", orderID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
		&order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee)

	if err != nil {
		return models.Order{}, fmt.Errorf("cannot scan info from Payments err:%v", err)
	}

	rows, err := pg.db.QueryContext(ctx, "SELECT * FROM Items WHERE order_id = $1", orderID)
	if err != nil {
		return models.Order{}, fmt.Errorf("cannot get info about items err:%v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var item models.Item

		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
			return models.Order{}, fmt.Errorf("cannot scan info item err:%v", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return models.Order{}, fmt.Errorf("error while scanning rows err:%v", err)
	}

	order.Items = items

	return order, nil

}

func (pg *PGStorage) GetAllOrderID(ctx context.Context) ([]string, error) {

	var IDs []string

	rows, err := pg.db.QueryContext(ctx, "SELECT order_uid FROM Orders")
	if err != nil {
		return nil, fmt.Errorf("cannot get all orderID from Orders err:%v", err)
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
		return nil, fmt.Errorf("error while scanning rows err:%v", err)
	}

	return IDs, nil

}
