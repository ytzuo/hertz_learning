package handler

import (
	"context"

	"Hertz/biz/model"
	"Hertz/biz/service"
	"Hertz/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req model.LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	login, err := h.authService.Login(ctx, req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, login)
}
