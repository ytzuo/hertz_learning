package handler

import (
	"context"

	"Hertz/biz/model"
	"Hertz/biz/service"
	"Hertz/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) List(ctx context.Context, c *app.RequestContext) {
	products, err := h.productService.List(ctx)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, products)
}

func (h *ProductHandler) Get(ctx context.Context, c *app.RequestContext) {
	product, err := h.productService.Get(ctx, c.Param("sku"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, product)
}

func (h *ProductHandler) AdjustStock(ctx context.Context, c *app.RequestContext) {
	var req model.AdjustStockRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	result, err := h.productService.AdjustStock(ctx, c.Param("sku"), req)
	if err != nil {
		response.BadRequest(c, err)
		return
	}

	response.OK(c, result)
}
