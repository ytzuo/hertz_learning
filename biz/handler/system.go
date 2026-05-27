package handler

import (
	"context"

	"Hertz/biz/model"
	"Hertz/biz/service"
	"Hertz/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type SystemHandler struct {
	systemService *service.SystemService
}

func NewSystemHandler(systemService *service.SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

func (h *SystemHandler) Health(ctx context.Context, c *app.RequestContext) {
	response.OK(c, h.systemService.Health(ctx))
}

func NoRoute(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusNotFound, model.ErrorResponse{
		Error: "route not found",
	})
}
