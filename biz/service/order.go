package service

import (
	"context"
	"errors"

	"Hertz/biz/model"
	"Hertz/infra/database"
	"Hertz/infra/mq"
)

var ErrForbidden = errors.New("forbidden")

type OrderService struct {
	db OrderRepository
	mq EventPublisher
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID string, items []database.OrderItem) (database.Order, error)
	GetOrder(ctx context.Context, orderID string) (database.Order, error)
	MarkOrderPaid(ctx context.Context, orderID string) (database.Order, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, event mq.Event) error
}

func NewOrderService(db OrderRepository, mq EventPublisher) *OrderService {
	return &OrderService{
		db: db,
		mq: mq,
	}
}

func (s *OrderService) Create(ctx context.Context, userID string, req model.CreateOrderRequest) (model.CreateOrderResponse, error) {
	items := make([]database.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, database.OrderItem{
			SKU: item.SKU,
			Qty: item.Qty,
		})
	}

	order, err := s.db.CreateOrder(ctx, userID, items)
	if err != nil {
		return model.CreateOrderResponse{}, err
	}

	// 状态变更成功后发布领域事件。
	// 真实系统通常会结合 outbox 表提升事件投递可靠性。
	_ = s.mq.Publish(ctx, mq.Event{
		Topic: "order.created",
		Key:   order.ID,
		Payload: map[string]any{
			"order_id": order.ID,
			"user_id":  order.UserID,
			"amount":   order.TotalAmount,
		},
	})

	return model.CreateOrderResponse{Order: toOrderResponse(order)}, nil
}

func (s *OrderService) Get(ctx context.Context, userID string, orderID string) (model.OrderResponse, error) {
	order, err := s.db.GetOrder(ctx, orderID)
	if err != nil {
		return model.OrderResponse{}, err
	}
	if order.UserID != userID {
		return model.OrderResponse{}, ErrForbidden
	}

	return toOrderResponse(order), nil
}

func (s *OrderService) Pay(ctx context.Context, userID string, orderID string) (model.PayOrderResponse, error) {
	order, err := s.db.GetOrder(ctx, orderID)
	if err != nil {
		return model.PayOrderResponse{}, err
	}
	if order.UserID != userID {
		return model.PayOrderResponse{}, ErrForbidden
	}

	paid, err := s.db.MarkOrderPaid(ctx, orderID)
	if err != nil {
		return model.PayOrderResponse{}, err
	}

	// 支付完成后向下游工作流分发事件，例如发票、通知、履约、数据分析等。
	_ = s.mq.Publish(ctx, mq.Event{
		Topic: "order.paid",
		Key:   paid.ID,
		Payload: map[string]any{
			"order_id": paid.ID,
			"user_id":  paid.UserID,
			"amount":   paid.TotalAmount,
		},
	})

	return model.PayOrderResponse{Order: toOrderResponse(paid)}, nil
}

func toOrderResponse(order database.Order) model.OrderResponse {
	items := make([]model.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, model.OrderItemResponse{
			SKU:       item.SKU,
			Name:      item.Name,
			UnitPrice: item.UnitPrice,
			Qty:       item.Qty,
			Amount:    item.Amount,
		})
	}

	resp := model.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Status:      order.Status,
		Items:       items,
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt.Unix(),
	}
	if !order.PaidAt.IsZero() {
		resp.PaidAt = order.PaidAt.Unix()
	}

	return resp
}
