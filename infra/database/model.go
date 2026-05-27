package database

import "time"

type User struct {
	ID           string `gorm:"primaryKey;size:64"`
	Name         string `gorm:"size:128"`
	Email        string `gorm:"uniqueIndex;size:255"`
	PasswordHash string `gorm:"size:255"`
	Role         string `gorm:"size:64"`
}

type Product struct {
	SKU       string `gorm:"primaryKey;size:64"`
	Name      string `gorm:"size:255"`
	Price     int
	Inventory int
}

type Order struct {
	ID          string `gorm:"primaryKey;size:64"`
	UserID      string `gorm:"index;size:64"`
	Status      string `gorm:"size:32"`
	Items       []OrderItem
	TotalAmount int
	CreatedAt   time.Time
	PaidAt      time.Time
}

type OrderItem struct {
	ID        uint   `gorm:"primaryKey"`
	OrderID   string `gorm:"index;size:64"`
	SKU       string
	Name      string
	UnitPrice int
	Qty       int
	Amount    int
}
