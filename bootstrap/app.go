package bootstrap

import (
	"context"
	"fmt"
	"strings"

	"Hertz/biz/handler"
	"Hertz/biz/router"
	"Hertz/biz/service"
	"Hertz/config"
	"Hertz/infra/cache"
	"Hertz/infra/database"
	"Hertz/infra/mq"
	"Hertz/pkg/auth"

	"github.com/cloudwego/hertz/pkg/app/server"
)

type App struct {
	h *server.Hertz
}

// NewApp 是当前服务实例的组合根。
// 真实项目通常也在这一层初始化配置、数据库连接池、Redis 客户端、
// MQ 生产者、Service、Handler 和 Router。
func NewApp() *App {
	cfg := config.Load()

	db, redis, eventBus := buildInfra(cfg)

	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.Issuer,
	)

	authService := service.NewAuthService(db, jwtManager)
	productService := service.NewProductService(db, redis)
	orderService := service.NewOrderService(db, eventBus)
	systemService := service.NewSystemService(cfg.Service)

	// Hertz 接管 HTTP 请求生命周期；业务依赖仍然由应用代码显式装配，
	// 而不是像 IoC 容器那样自动扫描。
	h := server.Default(server.WithHostPorts(cfg.HTTP.Addr))
	router.Register(h, router.Dependencies{
		SystemHandler:  handler.NewSystemHandler(systemService),
		AuthHandler:    handler.NewAuthHandler(authService),
		ProductHandler: handler.NewProductHandler(productService),
		OrderHandler:   handler.NewOrderHandler(orderService),
		JWTManager:     jwtManager,
		RateLimit:      cfg.RateLimit,
	})

	return &App{
		h: h,
	}
}

func (a *App) Run() {
	a.h.Spin()
}

type dataStore interface {
	service.UserRepository
	service.ProductRepository
	service.OrderRepository
}

func buildInfra(cfg config.Config) (dataStore, service.ProductCache, service.EventPublisher) {
	switch strings.ToLower(cfg.Infra.Adapter) {
	case "real":
		ctx := context.Background()

		db, err := database.NewMySQL(database.MySQLConfig{
			DSN: cfg.Infra.Database.DSN,
		})
		if err != nil {
			panic(fmt.Errorf("connect mysql: %w", err))
		}

		redis, err := cache.NewRedis(ctx, cache.RedisConfig{
			Addr:     cfg.Infra.Redis.Addr,
			Password: cfg.Infra.Redis.Password,
			DB:       cfg.Infra.Redis.DB,
		})
		if err != nil {
			panic(fmt.Errorf("connect redis: %w", err))
		}

		eventBus := mq.NewKafka(mq.KafkaConfig{
			Brokers: cfg.Infra.MQ.Brokers,
		})

		return db, redis, eventBus
	case "memory", "":
		db := database.NewMemoryDB()
		redis := cache.NewMemoryRedis()
		eventBus := mq.NewMemoryMQ()
		registerConsumers(eventBus)
		return db, redis, eventBus
	default:
		panic(fmt.Errorf("unsupported APP_ADAPTER %q", cfg.Infra.Adapter))
	}
}

// registerConsumers 模拟当前服务进程内的异步消费者。
// 生产项目里通常会放到 worker 包中，并从真实 MQ 消费消息。
func registerConsumers(eventBus *mq.MemoryMQ) {
	eventBus.Subscribe("order.created", func(ctx context.Context, event mq.Event) {
		fmt.Printf("consumer order.created reserve workflow key=%s payload=%v\n", event.Key, event.Payload)
	})
	eventBus.Subscribe("order.paid", func(ctx context.Context, event mq.Event) {
		fmt.Printf("consumer order.paid invoice workflow key=%s payload=%v\n", event.Key, event.Payload)
		fmt.Printf("consumer order.paid notify workflow key=%s payload=%v\n", event.Key, event.Payload)
	})
}
