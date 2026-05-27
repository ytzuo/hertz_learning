package model

type CreateOrderRequest struct {
	Items []CreateOrderItem `json:"items"`
}

type CreateOrderItem struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type CreateOrderResponse struct {
	Order OrderResponse `json:"order"`
}

type PayOrderResponse struct {
	Order OrderResponse `json:"order"`
}

type OrderResponse struct {
	ID          string              `json:"id"`
	UserID      string              `json:"user_id"`
	Status      string              `json:"status"`
	Items       []OrderItemResponse `json:"items"`
	TotalAmount int                 `json:"total_amount"`
	CreatedAt   int64               `json:"created_at"`
	PaidAt      int64               `json:"paid_at,omitempty"`
}

type OrderItemResponse struct {
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	UnitPrice int    `json:"unit_price"`
	Qty       int    `json:"qty"`
	Amount    int    `json:"amount"`
}
