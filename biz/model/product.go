package model

type ProductResponse struct {
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
	Inventory int    `json:"inventory"`
}

type AdjustStockRequest struct {
	Delta  int    `json:"delta"`
	Reason string `json:"reason"`
}

type AdjustStockResponse struct {
	SKU       string `json:"sku"`
	Inventory int    `json:"inventory"`
}
