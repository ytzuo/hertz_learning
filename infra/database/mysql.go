package database

import (
	"context"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
