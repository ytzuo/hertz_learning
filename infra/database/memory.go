package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrNotFound       = errors.New("record not found")
	ErrStockNotEnough = errors.New("stock not enough")
)

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

// MemoryDB 是 demo 使用的数据库适配器。
// 扣减库存、创建订单等类似事务的操作会放在同一个锁保护的临界区中。
type MemoryDB struct {
	mu       sync.RWMutex
	users    map[string]User
	products map[string]Product
	orders   map[string]Order
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		users: map[string]User{
			"admin@example.com": {
				ID:           "user-10001",
				Name:         "admin",
				Email:        "admin@example.com",
				PasswordHash: "password123",
				Role:         "admin",
			},
			"buyer@example.com": {
				ID:           "user-10002",
				Name:         "buyer",
				Email:        "buyer@example.com",
				PasswordHash: "password123",
				Role:         "buyer",
			},
		},
		products: map[string]Product{
			"book-go": {
				SKU:       "book-go",
				Name:      "Go Web Development",
				Price:     8900,
				Inventory: 20,
			},
			"keyboard": {
				SKU:       "keyboard",
				Name:      "Mechanical Keyboard",
				Price:     29900,
				Inventory: 5,
			},
			"mouse": {
				SKU:       "mouse",
				Name:      "Wireless Mouse",
				Price:     9900,
				Inventory: 10,
			},
		},
		orders: map[string]Order{},
	}
}

func (db *MemoryDB) FindUserByEmail(ctx context.Context, email string) (User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, ok := db.users[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (db *MemoryDB) ListProducts(ctx context.Context) ([]Product, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	products := make([]Product, 0, len(db.products))
	for _, product := range db.products {
		products = append(products, product)
	}
	return products, nil
}

func (db *MemoryDB) GetProduct(ctx context.Context, sku string) (Product, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	product, ok := db.products[sku]
	if !ok {
		return Product{}, ErrNotFound
	}
	return product, nil
}

func (db *MemoryDB) AdjustStock(ctx context.Context, sku string, delta int) (Product, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	product, ok := db.products[sku]
	if !ok {
		return Product{}, ErrNotFound
	}
	if product.Inventory+delta < 0 {
		return Product{}, ErrStockNotEnough
	}

	product.Inventory += delta
	db.products[sku] = product
	return product, nil
}

func (db *MemoryDB) CreateOrder(ctx context.Context, userID string, items []OrderItem) (Order, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if len(items) == 0 {
		return Order{}, errors.New("order items cannot be empty")
	}

	var total int
	reserved := make([]OrderItem, 0, len(items))
	// 这段逻辑模拟一个数据库事务：校验商品、预留库存、计算金额、写入订单。
	for _, item := range items {
		product, ok := db.products[item.SKU]
		if !ok {
			return Order{}, fmt.Errorf("product %q not found", item.SKU)
		}
		if item.Qty <= 0 {
			return Order{}, errors.New("item quantity must be greater than zero")
		}
		if product.Inventory < item.Qty {
			return Order{}, fmt.Errorf("product %q stock not enough", item.SKU)
		}

		product.Inventory -= item.Qty
		db.products[item.SKU] = product

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
	order := Order{
		ID:          fmt.Sprintf("ord-%d", now.UnixNano()),
		UserID:      userID,
		Status:      "created",
		Items:       reserved,
		TotalAmount: total,
		CreatedAt:   now,
	}
	db.orders[order.ID] = order

	return order, nil
}

func (db *MemoryDB) GetOrder(ctx context.Context, orderID string) (Order, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	order, ok := db.orders[orderID]
	if !ok {
		return Order{}, ErrNotFound
	}
	return order, nil
}

func (db *MemoryDB) MarkOrderPaid(ctx context.Context, orderID string) (Order, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	order, ok := db.orders[orderID]
	if !ok {
		return Order{}, ErrNotFound
	}
	if order.Status == "paid" {
		return order, nil
	}
	if order.Status != "created" {
		return Order{}, fmt.Errorf("order %s cannot be paid from status %s", orderID, order.Status)
	}

	order.Status = "paid"
	order.PaidAt = time.Now()
	db.orders[order.ID] = order
	return order, nil
}
