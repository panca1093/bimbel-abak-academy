package service

import (
	"context"
	"errors"
	"time"

	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/repository"
)

var (
	ErrForbidden       = errors.New("forbidden")
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidPromo    = errors.New("invalid or expired promo code")
	ErrPromoMinOrder   = errors.New("order subtotal below promo minimum")
)

type PromoValidation struct {
	Code     string
	Discount float64
	Total    float64
}

func (s *Service) ListProducts(ctx context.Context, filter repository.ProductFilter, role string) ([]repository.Product, string, error) {
	switch role {
	case RoleSuperAdmin:
		// no filter restrictions
	case RoleAdminStore:
		if filter.Type == "exam" {
			return nil, "", nil
		}
	case RoleAdminExam:
		if filter.Type != "" && filter.Type != "exam" {
			return nil, "", nil
		}
		filter.Type = "exam"
	default: // student or ""
		filter.IsVisibleOnly = true
		filter.Status = "published"
	}
	return s.storeRepo.ListProducts(ctx, filter)
}

func (s *Service) GetProduct(ctx context.Context, id string, role string) (repository.Product, error) {
	p, err := s.storeRepo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.Product{}, ErrProductNotFound
		}
		return repository.Product{}, err
	}
	if role == RoleStudent || role == "" {
		if p.Status != "published" || !p.IsVisible {
			return repository.Product{}, ErrProductNotFound
		}
	}
	return *p, nil
}

func (s *Service) CreateProduct(ctx context.Context, p repository.Product, role string) (repository.Product, error) {
	if err := checkTypeRBAC(role, p.Type); err != nil {
		return repository.Product{}, err
	}
	if err := s.storeRepo.CreateProduct(ctx, &p); err != nil {
		return repository.Product{}, err
	}
	return p, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id string, p repository.Product, role string) (repository.Product, error) {
	existing, err := s.storeRepo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.Product{}, ErrProductNotFound
		}
		return repository.Product{}, err
	}
	if err := checkTypeRBAC(role, existing.Type); err != nil {
		return repository.Product{}, err
	}
	if err := s.storeRepo.UpdateProduct(ctx, id, &p); err != nil {
		return repository.Product{}, err
	}
	p.ID = id
	return p, nil
}

func (s *Service) PublishProduct(ctx context.Context, id string, role string) error {
	existing, err := s.storeRepo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	if err := checkTypeRBAC(role, existing.Type); err != nil {
		return err
	}
	return s.storeRepo.PublishProduct(ctx, id)
}

func (s *Service) DeleteProduct(ctx context.Context, id string, role string) error {
	existing, err := s.storeRepo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	if err := checkTypeRBAC(role, existing.Type); err != nil {
		return err
	}
	if existing.Type == "book" {
		return s.storeRepo.DeleteProduct(ctx, id)
	}
	return s.storeRepo.ArchiveProduct(ctx, id)
}

func (s *Service) ValidatePromo(ctx context.Context, code string, subtotal float64) (PromoValidation, error) {
	promo, err := s.storeRepo.GetPromoByCode(ctx, code)
	if err != nil {
		return PromoValidation{}, err
	}
	if promo.Code == "" {
		return PromoValidation{}, ErrInvalidPromo
	}
	if promo.ExpiresAt != nil && promo.ExpiresAt.Before(time.Now()) {
		return PromoValidation{}, ErrInvalidPromo
	}
	if promo.MaxUses != nil && promo.Uses >= *promo.MaxUses {
		return PromoValidation{}, ErrInvalidPromo
	}
	if promo.MinOrderAmount != nil && subtotal < *promo.MinOrderAmount {
		return PromoValidation{}, ErrPromoMinOrder
	}

	var discount float64
	if promo.DiscountPercent != nil {
		discount = subtotal * (*promo.DiscountPercent / 100)
		if promo.MaxDiscountAmount != nil && discount > *promo.MaxDiscountAmount {
			discount = *promo.MaxDiscountAmount
		}
	} else if promo.DiscountAmount != nil {
		discount = *promo.DiscountAmount
		if discount > subtotal {
			discount = subtotal
		}
	}

	return PromoValidation{Code: code, Discount: discount, Total: subtotal - discount}, nil
}

func (s *Service) GetShippingRates(ctx context.Context, req platform.ShippingQuoteRequest) ([]platform.CourierRate, error) {
	return s.logistics.GetRates(ctx, req)
}

// checkTypeRBAC returns ErrForbidden if role is not allowed to manage productType.
func checkTypeRBAC(role, productType string) error {
	switch role {
	case RoleSuperAdmin:
		return nil
	case RoleAdminStore:
		if productType == "book" || productType == "course" {
			return nil
		}
		return ErrForbidden
	case RoleAdminExam:
		if productType == "exam" {
			return nil
		}
		return ErrForbidden
	default:
		return ErrForbidden
	}
}
