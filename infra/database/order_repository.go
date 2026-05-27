package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (m *MySQLDB) CreateOrder(ctx context.Context, userID string, items []OrderItem) (Order, error) {
	var order Order
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(items) == 0 {
			return errors.New("order items cannot be empty")
		}

		total := 0
		reserved := make([]OrderItem, 0, len(items))
		for _, item := range items {
			if item.Qty <= 0 {
				return errors.New("item quantity must be greater than zero")
			}

			var product Product
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sku = ?", item.SKU).First(&product).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("product %q not found", item.SKU)
				}
				return err
			}
			if product.Inventory < item.Qty {
				return fmt.Errorf("product %q stock not enough", item.SKU)
			}

			product.Inventory -= item.Qty
			if err := tx.Save(&product).Error; err != nil {
				return err
			}

			amount := product.Price * item.Qty
			total += amount
			reserved = append(reserved, OrderItem{
				SKU:       product.SKU,
				Name:      product.Name,
				UnitPrice: product.Price,
				Qty:       item.Qty,
				Amount:    amount,
			})
		}

		now := time.Now()
		order = Order{
			ID:          fmt.Sprintf("ord-%d", now.UnixNano()),
			UserID:      userID,
			Status:      "created",
			Items:       reserved,
			TotalAmount: total,
			CreatedAt:   now,
		}
		return tx.Create(&order).Error
	})
	return order, err
}

func (m *MySQLDB) GetOrder(ctx context.Context, orderID string) (Order, error) {
	var order Order
	err := m.db.WithContext(ctx).Preload("Items").Where("id = ?", orderID).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Order{}, ErrNotFound
	}
	return order, err
}

func (m *MySQLDB) MarkOrderPaid(ctx context.Context, orderID string) (Order, error) {
	var order Order
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").Where("id = ?", orderID).First(&order).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if order.Status == "paid" {
			return nil
		}
		if order.Status != "created" {
			return fmt.Errorf("order %s cannot be paid from status %s", orderID, order.Status)
		}

		order.Status = "paid"
		order.PaidAt = time.Now()
		return tx.Save(&order).Error
	})
	return order, err
}
