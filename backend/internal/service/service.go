package service

import (
	"context"

	"akademi-bimbel/internal/config"
	"akademi-bimbel/internal/platform"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo          UserRepository
	rdb           *redis.Client
	jwtSigner     *platform.JWTSigner
	otpProvider   platform.OTPProvider
	emailProvider platform.EmailProvider
	cfg           *config.Config
}

// NewForTest builds a Service with only a Redis client — sufficient for middleware tests.
func NewForTest(rdb *redis.Client) *Service {
	return &Service{rdb: rdb}
}

func New(
	repo UserRepository,
	rdb *redis.Client,
	jwtSigner *platform.JWTSigner,
	otpProvider platform.OTPProvider,
	emailProvider platform.EmailProvider,
	cfg *config.Config,
) *Service {
	return &Service{
		repo:          repo,
		rdb:           rdb,
		jwtSigner:     jwtSigner,
		otpProvider:   otpProvider,
		emailProvider: emailProvider,
		cfg:           cfg,
	}
}

type Health struct {
	Status   string `json:"status"`
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
}

func (s *Service) ParseAccess(tokenString string) (*platform.Claims, error) {
	return s.jwtSigner.ParseAccess(tokenString)
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
