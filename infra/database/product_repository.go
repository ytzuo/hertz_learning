package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
