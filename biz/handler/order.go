package handler

import (
	"context"

	"Hertz/biz/middleware"
	"Hertz/biz/model"
	"Hertz/biz/service"
	"Hertz/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) Create(ctx context.Context, c *app.RequestContext) {
	var req model.CreateOrderRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	result, err := h.orderService.Create(ctx, middleware.CurrentUserID(c), req)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.Created(c, result)
}

func (h *OrderHandler) Get(ctx context.Context, c *app.RequestContext) {
	order, err := h.orderService.Get(ctx, middleware.CurrentUserID(c), c.Param("id"))
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.OK(c, order)
}

func (h *OrderHandler) Pay(ctx context.Context, c *app.RequestContext) {
	result, err := h.orderService.Pay(ctx, middleware.CurrentUserID(c), c.Param("id"))
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.OK(c, result)
}
