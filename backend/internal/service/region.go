package service

import (
	"context"

	"akademi-bimbel/internal/model"
)

// ListProvinces returns all provinces ordered alphabetically.
func (s *Service) ListProvinces(ctx context.Context) ([]model.Province, error) {
	return s.storeRepo.ListProvinces(ctx)
}

// ListCitiesByProvince returns cities for the given province id.
// Returns an empty slice (not an error) when the province id is unknown.
func (s *Service) ListCitiesByProvince(ctx context.Context, provinceID string) ([]model.City, error) {
	return s.storeRepo.ListCitiesByProvince(ctx, provinceID)
}

// ListDistrictsByCity returns districts for the given city id.
// Returns an empty slice (not an error) when the city id is unknown.
func (s *Service) ListDistrictsByCity(ctx context.Context, cityID string) ([]model.District, error) {
	return s.storeRepo.ListDistrictsByCity(ctx, cityID)
}
