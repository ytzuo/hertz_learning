package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

func AccessLog() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		c.Next(ctx)

		requestID, _ := c.Get(RequestIDKey)
		userID, _ := c.Get(UserIDKey)

		fmt.Printf("request_id=%v user_id=%v method=%s path=%s status=%d cost=%s\n",
			requestID,
			userID,
			c.Method(),
			c.Path(),
			c.Response.StatusCode(),
			time.Since(start),
		)
	}
}
