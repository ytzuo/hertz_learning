package database

import "errors"

var (
	ErrNotFound       = errors.New("record not found")
	ErrStockNotEnough = errors.New("stock not enough")
)
