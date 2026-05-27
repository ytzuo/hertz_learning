package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MySQLConfig struct {
	DSN string
}

type MySQLDB struct {
	db *gorm.DB
}

func NewMySQL(cfg MySQLConfig) (*MySQLDB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return &MySQLDB{db: db}, nil
}

func (m *MySQLDB) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (m *MySQLDB) AutoMigrate(ctx context.Context) error {
	return m.db.WithContext(ctx).AutoMigrate(&User{}, &Product{}, &Order{}, &OrderItem{})
}

func (m *MySQLDB) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := m.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (m *MySQLDB) ListProducts(ctx context.Context) ([]Product, error) {
	var products []Product
	err := m.db.WithContext(ctx).Find(&products).Error
	return products, err
}

func (m *MySQLDB) GetProduct(ctx context.Context, sku string) (Product, error) {
	var product Product
	err := m.db.WithContext(ctx).Where("sku = ?", sku).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Product{}, ErrNotFound
	}
	return product, err
}

func (m *MySQLDB) AdjustStock(ctx context.Context, sku string, delta int) (Product, error) {
	var product Product
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sku = ?", sku).First(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if product.Inventory+delta < 0 {
			return ErrStockNotEnough
		}

		product.Inventory += delta
		return tx.Save(&product).Error
	})
	return product, err
}

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
