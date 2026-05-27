package middleware

import (
	"context"
	"fmt"
	"strings"

	"Hertz/pkg/auth"
	"Hertz/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

const UserIDKey = "user_id"
const UserEmailKey = "user_email"
const UserRoleKey = "user_role"

// Auth 把 Bearer JWT 转换成当前请求范围内的用户身份。
// 下游 handler 应从 RequestContext 读取身份，再把普通参数传给 service，
// 不要把 Hertz 的 RequestContext 继续传入业务层。
func Auth(jwtManager *auth.JWTManager) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		authorization := strings.TrimSpace(string(c.GetHeader("Authorization")))
		if authorization == "" {
			response.Unauthorized(c, "missing authorization token")
			c.Abort()
			return
		}

		token, err := auth.BearerToken(authorization)
		if err != nil {
			response.Unauthorized(c, "invalid authorization token")
			c.Abort()
			return
		}

		claims, err := jwtManager.Verify(token)
		if err != nil {
			response.Unauthorized(c, "invalid authorization token")
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.Subject)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UserRoleKey, roleFromEmail(claims.Email))
		c.Next(ctx)
	}
}

// RequireRole 是路由级授权策略。
// 它依赖 Auth 先执行并写入 UserRoleKey。
func RequireRole(role string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		current, ok := c.Get(UserRoleKey)
		if !ok || fmt.Sprint(current) != role {
			response.Forbidden(c, "forbidden")
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

// CurrentUserID 是给 handler 使用的 HTTP 层辅助函数。
func CurrentUserID(c *app.RequestContext) string {
	userID, ok := c.Get(UserIDKey)
	if !ok {
		return ""
	}
	return fmt.Sprint(userID)
}

func roleFromEmail(email string) string {
	if email == "admin@example.com" {
		return "admin"
	}
	return "buyer"
}
