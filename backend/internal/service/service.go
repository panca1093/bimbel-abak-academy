package service

import (
	"context"

	"akademi-bimbel/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo *repository.Repository
	rdb  *redis.Client
}

func New(repo *repository.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

type Health struct {
	Status   string `json:"status"`
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
}

func (s *Service) Health(ctx context.Context) Health {
	h := Health{Status: "ok", Postgres: "ok", Redis: "ok"}
	if err := s.repo.Ping(ctx); err != nil {
		h.Postgres = "down"
		h.Status = "degraded"
	}
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		h.Redis = "down"
		h.Status = "degraded"
	}
	return h
}
