package router

import (
	"Hertz/biz/handler"
	"Hertz/biz/middleware"
	"Hertz/config"
	"Hertz/pkg/auth"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"
)

type Dependencies struct {
	SystemHandler  *handler.SystemHandler
	AuthHandler    *handler.AuthHandler
	ProductHandler *handler.ProductHandler
	OrderHandler   *handler.OrderHandler
	JWTManager     *auth.JWTManager
	RateLimit      config.RateLimitConfig
}

// RouteSpec 显式描述每个接口的策略。
// 这样不需要为了鉴权、限流、角色等小差异创建大量 route group。
type RouteSpec struct {
	Method   string
	Path     string
	Policies []app.HandlerFunc
	Handler  app.HandlerFunc
}

func Register(h *server.Hertz, deps Dependencies) {
	registerGlobalMiddleware(h)

	h.GET("/", deps.SystemHandler.Health)
	h.GET("/health", deps.SystemHandler.Health)

	api := h.Group("/api/v1")
	registerRoutes(api, buildRoutes(deps))

	h.NoRoute(handler.NoRoute)
}

func registerGlobalMiddleware(h *server.Hertz) {
	h.Use(
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.AccessLog(),
	)
}

func buildRoutes(deps Dependencies) []RouteSpec {
	// policy 本质上是一段可复用的 Hertz handler chain 前缀。
	// 最终业务 handler 会在 registerRoutes 中被追加到链条最后。
	publicRead := []app.HandlerFunc{
		middleware.RateLimit(deps.RateLimit.PublicMaxRequests, deps.RateLimit.Window),
	}
	privateRead := []app.HandlerFunc{
		middleware.Auth(deps.JWTManager),
		middleware.RateLimit(deps.RateLimit.PrivateMaxRequests, deps.RateLimit.Window),
	}
	orderWrite := []app.HandlerFunc{
		middleware.Auth(deps.JWTManager),
		middleware.RateLimit(deps.RateLimit.OrderMaxRequests, deps.RateLimit.Window),
	}
	adminWrite := []app.HandlerFunc{
		middleware.Auth(deps.JWTManager),
		middleware.RequireRole("admin"),
		middleware.RateLimit(deps.RateLimit.PrivateMaxRequests, deps.RateLimit.Window),
	}

	return []RouteSpec{
		{Method: "POST", Path: "/auth/login", Policies: publicRead, Handler: deps.AuthHandler.Login},
		{Method: "GET", Path: "/products", Policies: publicRead, Handler: deps.ProductHandler.List},
		{Method: "GET", Path: "/products/:sku", Policies: publicRead, Handler: deps.ProductHandler.Get},
		{Method: "POST", Path: "/products/:sku/stock", Policies: adminWrite, Handler: deps.ProductHandler.AdjustStock},
		{Method: "POST", Path: "/orders", Policies: orderWrite, Handler: deps.OrderHandler.Create},
		{Method: "GET", Path: "/orders/:id", Policies: privateRead, Handler: deps.OrderHandler.Get},
		{Method: "POST", Path: "/orders/:id/pay", Policies: orderWrite, Handler: deps.OrderHandler.Pay},
	}
}

func registerRoutes(group *route.RouterGroup, routes []RouteSpec) {
	for _, spec := range routes {
		// Hertz 允许一个路由绑定多个 handler。
		// 路由级策略先执行，最终的接口 handler 最后执行。
		handlers := append([]app.HandlerFunc{}, spec.Policies...)
		handlers = append(handlers, spec.Handler)

		switch spec.Method {
		case "GET":
			group.GET(spec.Path, handlers...)
		case "POST":
			group.POST(spec.Path, handlers...)
		}
	}
}

func Window(max int, window time.Duration) config.RateLimitConfig {
	return config.RateLimitConfig{
		PublicMaxRequests:  max,
		PrivateMaxRequests: max,
		OrderMaxRequests:   max,
		Window:             window,
	}
}
