# Hertz 商城 API Demo

这是一个单体微服务实例形态的 Hertz Demo。它不是只演示几个路由，而是模拟一个比较真实的业务服务：有登录认证、商品库存、订单下单、支付、缓存、消息事件和接口级策略。

当前项目可以直接本地运行，不依赖真实 MySQL、Redis、MQ。相关中间件先用本地适配器模拟，方便学习业务流程和项目组织方式。

## 业务模块

- 认证业务：登录、JWT 签发、JWT 认证。
- 商品业务：商品列表、商品详情、商品详情缓存、管理员调整库存。
- 订单业务：创建订单、扣减库存、查询订单、支付订单、发布订单事件。

## 中间件模拟

- `infra/database`：内存数据库，模拟用户、商品、订单和库存扣减。真实项目里可以替换成 MySQL/PostgreSQL repository。
- `infra/cache`：内存 Redis，商品详情使用 cache-aside 模式。
- `infra/mq`：内存 MQ，订单创建和支付后发布事件，并由本进程内消费者异步处理。

## 请求链路

```text
HTTP 请求
  -> Hertz Router
  -> 全局中间件：Recovery / RequestID / AccessLog
  -> 路由级策略：Auth / RateLimit / RequireRole
  -> Handler：绑定参数、调用 Service、写响应
  -> Service：编排业务流程
  -> Infra：数据库、Redis 缓存、MQ 事件总线
```

## 目录结构

```text
.
├── bootstrap      # 应用装配入口，初始化配置、DB、Redis、MQ、Service、Handler、Router
├── biz
│   ├── handler    # HTTP 适配层，只处理入参、出参和响应
│   ├── middleware # Hertz 中间件：认证、限流、日志、恢复、请求 ID
│   ├── model      # 请求和响应模型
│   ├── router     # 路由注册和接口级策略组合
│   └── service    # 业务逻辑和业务流程编排
├── config         # 本地配置
├── infra
│   ├── cache      # Redis-like 适配器
│   ├── database   # DB-like 适配器
│   └── mq         # MQ-like 适配器
└── pkg
    ├── auth       # JWT 签发和校验
    └── response   # 统一响应辅助
```

## 启动

```bash
go run .
```

服务默认监听：

```text
http://localhost:8888
```

健康检查：

```bash
curl http://localhost:8888/health
```

## 测试账号

普通买家：

```text
buyer@example.com / password123
```

管理员：

```text
admin@example.com / password123
```

## 接口示例

登录：

```bash
curl -X POST http://localhost:8888/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"buyer@example.com\",\"password\":\"password123\"}"
```

查询商品列表：

```bash
curl http://localhost:8888/api/v1/products
```

查询商品详情：

```bash
curl http://localhost:8888/api/v1/products/book-go
```

创建订单：

```bash
curl -X POST http://localhost:8888/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d "{\"items\":[{\"sku\":\"book-go\",\"qty\":1},{\"sku\":\"mouse\",\"qty\":2}]}"
```

查询订单：

```bash
curl http://localhost:8888/api/v1/orders/<order_id> \
  -H "Authorization: Bearer <access_token>"
```

支付订单：

```bash
curl -X POST http://localhost:8888/api/v1/orders/<order_id>/pay \
  -H "Authorization: Bearer <access_token>"
```

管理员调整库存：

```bash
curl -X POST http://localhost:8888/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"admin@example.com\",\"password\":\"password123\"}"

curl -X POST http://localhost:8888/api/v1/products/book-go/stock \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_access_token>" \
  -d "{\"delta\":10,\"reason\":\"manual replenish\"}"
```

## 设计说明

- `bootstrap` 是应用组合根，负责显式装配所有依赖。
- `router` 使用 `RouteSpec` 描述接口，给每个接口挂不同的认证、限流、角色策略，避免创建大量 group。
- `handler` 是薄层，只做 HTTP 协议适配，不承载复杂业务逻辑。
- `service` 是业务核心，负责订单创建、库存扣减、支付、发布事件等流程编排。
- `infra` 是可替换的基础设施适配层，当前用内存实现，真实项目中可以替换成实际中间件客户端。
- `middleware` 是 Hertz handler chain 的一部分，用来处理横切逻辑。

## 重点学习点

- Hertz 只调度 `app.HandlerFunc` 链。
- 中间件和业务 handler 本质都是 handler，只是所在位置和职责不同。
- 业务 service 不依赖 Hertz 的 `RequestContext`，只接收普通参数和 `context.Context`。
- 接口级差异不一定要拆很多 group，可以用路由策略组合表达。
- DB、Redis、MQ 应该放在基础设施层，通过 service 编排，而不是直接写进 handler。
