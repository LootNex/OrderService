package models

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"time"
)

type Validator interface {
	Validate() error
}

func (o *Order) Validate() error {
	if o.OrderUID == "" {
		return errors.New("order_uid is required")
	}
	if o.TrackNumber == "" {
		return errors.New("track_number is required")
	}
	if o.CustomerID == "" {
		return errors.New("customer_id is required")
	}
	if o.DeliveryService == "" {
		return errors.New("delivery_service is required")
	}

	if _, err := time.Parse(time.RFC3339, o.DateCreated); err != nil {
		return fmt.Errorf("invalid date_created format: %w", err)
	}

	if err := o.Delivery.Validate(); err != nil {
		return fmt.Errorf("delivery validation failed: %w", err)
	}
	if err := o.Payment.Validate(); err != nil {
		return fmt.Errorf("payment validation failed: %w", err)
	}
	if len(o.Items) == 0 {
		return errors.New("order must contain at least one item")
	}
	for i, item := range o.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("item[%d] validation failed: %w", i, err)
		}
	}

	return nil
}

func (d *Delivery) Validate() error {
	if d.Name == "" {
		return errors.New("delivery name is required")
	}
	if d.Phone == "" {
		return errors.New("delivery phone is required")
	}
	if _, err := mail.ParseAddress(d.Email); err != nil {
		return errors.New("invalid delivery email")
	}
	return nil
}

func (p *Payment) Validate() error {
	if p.Transaction == "" {
		return errors.New("transaction is required")
	}
	if p.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if p.Currency == "" {
		return errors.New("currency is required")
	}
	return nil
}

func (i *Item) Validate() error {
	if i.ChrtID == 0 {
		return errors.New("chrt_id is required")
	}
	if i.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if i.TotalPrice < 0 {
		return errors.New("total_price cannot be negative")
	}
	if matched, _ := regexp.MatchString(`^[A-Z0-9]+$`, i.TrackNumber); !matched {
		return errors.New("invalid track_number format in item")
	}
	return nil
}
