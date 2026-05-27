package middleware

import (
	"context"
	"fmt"

	"Hertz/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get(RequestIDKey)
				fmt.Printf("panic recovered request_id=%v error=%v\n", requestID, err)
				response.InternalError(c)
				c.Abort()
			}
		}()

		c.Next(ctx)
	}
}
