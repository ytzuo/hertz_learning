package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"Hertz/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type visitorBucket struct {
	count   int
	expires time.Time
}

func RateLimit(maxRequests int, window time.Duration) app.HandlerFunc {
	if maxRequests <= 0 {
		maxRequests = 60
	}
	if window <= 0 {
		window = time.Minute
	}

	var mu sync.Mutex
	visitors := map[string]visitorBucket{}

	return func(ctx context.Context, c *app.RequestContext) {
		now := time.Now()
		key := rateLimitKey(c)

		mu.Lock()
		bucket := visitors[key]
		if bucket.expires.Before(now) {
			bucket = visitorBucket{expires: now.Add(window)}
		}
		bucket.count++
		visitors[key] = bucket
		limited := bucket.count > maxRequests
		mu.Unlock()

		if limited {
			response.TooManyRequests(c, "too many requests")
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

func rateLimitKey(c *app.RequestContext) string {
	userID, ok := c.Get(UserIDKey)
	if ok {
		return "user:" + fmt.Sprint(userID)
	}

	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := string(c.GetHeader(header))
		if value != "" {
			return "ip:" + value
		}
	}

	return "ip:unknown"
}
