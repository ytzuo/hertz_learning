package service

import (
	"context"

	"Hertz/biz/model"
	"Hertz/config"
)

type SystemService struct {
	cfg config.ServiceConfig
}

func NewSystemService(cfg config.ServiceConfig) *SystemService {
	return &SystemService{cfg: cfg}
}

func (s *SystemService) Health(ctx context.Context) model.HealthResponse {
	return model.HealthResponse{
		Service: s.cfg.Name,
		Version: s.cfg.Version,
		Env:     s.cfg.Env,
		Status:  "ok",
	}
}
