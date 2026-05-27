package response

import (
	"Hertz/biz/model"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func OK(c *app.RequestContext, data any) {
	c.JSON(consts.StatusOK, data)
}

func Created(c *app.RequestContext, data any) {
	c.JSON(consts.StatusCreated, data)
}

func BadRequest(c *app.RequestContext, err error) {
	c.JSON(consts.StatusBadRequest, model.ErrorResponse{
		Error: err.Error(),
	})
}

func NotFound(c *app.RequestContext, message string) {
	c.JSON(consts.StatusNotFound, model.ErrorResponse{
		Error: message,
	})
}

func Unauthorized(c *app.RequestContext, message string) {
	c.JSON(consts.StatusUnauthorized, model.ErrorResponse{
		Error: message,
	})
}

func Forbidden(c *app.RequestContext, message string) {
	c.JSON(consts.StatusForbidden, model.ErrorResponse{
		Error: message,
	})
}

func TooManyRequests(c *app.RequestContext, message string) {
	c.JSON(consts.StatusTooManyRequests, model.ErrorResponse{
		Error: message,
	})
}

func InternalError(c *app.RequestContext) {
	c.JSON(consts.StatusInternalServerError, model.ErrorResponse{
		Error: "internal server error",
	})
}
