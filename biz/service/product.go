package service

import (
	"context"
	"fmt"
	"time"

	"Hertz/biz/model"
	"Hertz/infra/cache"
	"Hertz/infra/database"
)

type ProductService struct {
	db    *database.MemoryDB
	cache *cache.MemoryRedis
}

func NewProductService(db *database.MemoryDB, cache *cache.MemoryRedis) *ProductService {
	return &ProductService{
		db:    db,
		cache: cache,
	}
}

func (s *ProductService) List(ctx context.Context) ([]model.ProductResponse, error) {
	products, err := s.db.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]model.ProductResponse, 0, len(products))
	for _, product := range products {
		resp = append(resp, toProductResponse(product))
	}
	return resp, nil
}

func (s *ProductService) Get(ctx context.Context, sku string) (model.ProductResponse, error) {
	key := productCacheKey(sku)
	// Cache-aside 模式：先读缓存；缓存未命中时查数据库，然后回填缓存。
	if cached, err := s.cache.Get(ctx, key); err == nil {
		if product, ok := cached.(model.ProductResponse); ok {
			return product, nil
		}
	}

	product, err := s.db.GetProduct(ctx, sku)
	if err != nil {
		return model.ProductResponse{}, err
	}

	resp := toProductResponse(product)
	_ = s.cache.Set(ctx, key, resp, 30*time.Second)
	return resp, nil
}

func (s *ProductService) AdjustStock(ctx context.Context, sku string, req model.AdjustStockRequest) (model.AdjustStockResponse, error) {
	product, err := s.db.AdjustStock(ctx, sku, req.Delta)
	if err != nil {
		return model.AdjustStockResponse{}, err
	}

	// 库存变更后删除商品详情缓存，避免读到旧库存。
	_ = s.cache.Delete(ctx, productCacheKey(sku))
	return model.AdjustStockResponse{
		SKU:       product.SKU,
		Inventory: product.Inventory,
	}, nil
}

func toProductResponse(product database.Product) model.ProductResponse {
	return model.ProductResponse{
		SKU:       product.SKU,
		Name:      product.Name,
		Price:     product.Price,
		Inventory: product.Inventory,
	}
}

func productCacheKey(sku string) string {
	return fmt.Sprintf("product:%s", sku)
}
