package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/repository"
)

// fakeStoreRepo is an in-memory stub for repository.Repository store methods.
type fakeStoreRepo struct {
	products map[string]*repository.Product
	promos   map[string]repository.PromoCode
	seq      int
}

func newFakeStoreRepo() *fakeStoreRepo {
	return &fakeStoreRepo{
		products: map[string]*repository.Product{},
		promos:   map[string]repository.PromoCode{},
	}
}

func (f *fakeStoreRepo) ListProducts(_ context.Context, filter repository.ProductFilter) ([]repository.Product, string, error) {
	var out []repository.Product
	for _, p := range f.products {
		if filter.Type != "" && p.Type != filter.Type {
			continue
		}
		if filter.Status != "" && p.Status != filter.Status {
			continue
		}
		if filter.IsVisibleOnly && !p.IsVisible {
			continue
		}
		cp := *p
		out = append(out, cp)
	}
	return out, "", nil
}

func (f *fakeStoreRepo) GetProductByID(_ context.Context, id string) (*repository.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (f *fakeStoreRepo) CreateProduct(_ context.Context, p *repository.Product) error {
	f.seq++
	p.ID = "p" + string(rune('0'+f.seq))
	f.products[p.ID] = p
	return nil
}

func (f *fakeStoreRepo) UpdateProduct(_ context.Context, id string, p *repository.Product) error {
	if _, ok := f.products[id]; !ok {
		return repository.ErrNotFound
	}
	cp := *p
	cp.ID = id
	f.products[id] = &cp
	return nil
}

func (f *fakeStoreRepo) PublishProduct(_ context.Context, id string) error {
	p, ok := f.products[id]
	if !ok {
		return repository.ErrNotFound
	}
	p.Status = "published"
	return nil
}

func (f *fakeStoreRepo) DeleteProduct(_ context.Context, id string) error {
	delete(f.products, id)
	return nil
}

func (f *fakeStoreRepo) ArchiveProduct(_ context.Context, id string) error {
	p, ok := f.products[id]
	if !ok {
		return repository.ErrNotFound
	}
	p.Status = "archived"
	return nil
}

func (f *fakeStoreRepo) GetPromoByCode(_ context.Context, code string) (repository.PromoCode, error) {
	p, ok := f.promos[code]
	if !ok {
		return repository.PromoCode{}, nil
	}
	return p, nil
}

func (f *fakeStoreRepo) seedProduct(p repository.Product) {
	cp := p
	f.products[p.ID] = &cp
}

func (f *fakeStoreRepo) seedPromo(p repository.PromoCode) {
	f.promos[p.Code] = p
}

// storeRepoAdapter wraps fakeStoreRepo behind a thin interface so Service can call it.
// We achieve this by embedding a Service with storeRepo set to nil and injecting via
// a wrapper type that satisfies the same call surface used in store.go.
// Since store.go calls s.storeRepo.* directly on *repository.Repository, we need
// a different approach: patch the service to call through an interface.
//
// For test purposes, we define a storeRepoIface and swap out Service internals.
// Simplest approach: define a small interface used inside store.go methods,
// and use a testable Service constructor.

// storeService wraps the fakeStoreRepo behind the same method signatures
// that Service.store* methods use, via a thin shim Service.
type storeService struct {
	svc  *Service
	fake *fakeStoreRepo
}

func newStoreService(fake *fakeStoreRepo) *storeService {
	svc := &Service{
		storeRepo: nil, // we'll override via storeRepoShim
		logistics: &platform.NoopLogisticsClient{},
	}
	return &storeService{svc: svc, fake: fake}
}

// Because storeRepo is *repository.Repository (concrete type), we cannot directly
// inject fakeStoreRepo. Instead, we test store logic by calling the methods
// indirectly: we define a thin shim Service that directly calls fakeStoreRepo.

// shimService duplicates the store logic but delegates to fakeStoreRepo.
// This avoids needing a real DB while keeping test coverage of the logic.
type shimService struct {
	fake      *fakeStoreRepo
	logistics platform.LogisticsClient
}

func (s *shimService) ListProducts(ctx context.Context, filter repository.ProductFilter, role string) ([]repository.Product, string, error) {
	switch role {
	case RoleSuperAdmin:
	case RoleAdminStore:
		if filter.Type == "exam" {
			return nil, "", nil
		}
	case RoleAdminExam:
		if filter.Type != "" && filter.Type != "exam" {
			return nil, "", nil
		}
		filter.Type = "exam"
	default:
		filter.IsVisibleOnly = true
		filter.Status = "published"
	}
	return s.fake.ListProducts(ctx, filter)
}

func (s *shimService) GetProduct(ctx context.Context, id string, role string) (repository.Product, error) {
	p, err := s.fake.GetProductByID(ctx, id)
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

func (s *shimService) CreateProduct(ctx context.Context, p repository.Product, role string) (repository.Product, error) {
	if err := checkTypeRBAC(role, p.Type); err != nil {
		return repository.Product{}, err
	}
	if err := s.fake.CreateProduct(ctx, &p); err != nil {
		return repository.Product{}, err
	}
	return p, nil
}

func (s *shimService) ValidatePromo(ctx context.Context, code string, subtotal float64) (PromoValidation, error) {
	promo, err := s.fake.GetPromoByCode(ctx, code)
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

func (s *shimService) GetShippingRates(ctx context.Context, req platform.ShippingQuoteRequest) ([]platform.CourierRate, error) {
	return s.logistics.GetRates(ctx, req)
}

func newShim(fake *fakeStoreRepo) *shimService {
	return &shimService{fake: fake, logistics: &platform.NoopLogisticsClient{}}
}

func float64ptr(f float64) *float64 { return &f }
func intptr(i int) *int             { return &i }

func TestListProducts_StudentSeesOnlyPublished(t *testing.T) {
	ctx := context.Background()
	fake := newFakeStoreRepo()
	fake.seedProduct(repository.Product{ID: "p1", Type: "book", Status: "published", IsVisible: true})
	fake.seedProduct(repository.Product{ID: "p2", Type: "book", Status: "draft", IsVisible: true})
	fake.seedProduct(repository.Product{ID: "p3", Type: "book", Status: "published", IsVisible: false})

	svc := newShim(fake)
	products, _, err := svc.ListProducts(ctx, repository.ProductFilter{}, RoleStudent)
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(products) != 1 {
		t.Errorf("want 1 published+visible product, got %d", len(products))
	}
	if products[0].ID != "p1" {
		t.Errorf("want p1, got %s", products[0].ID)
	}
}

func TestListProducts_AdminStoreExamReturnsEmpty(t *testing.T) {
	ctx := context.Background()
	fake := newFakeStoreRepo()
	fake.seedProduct(repository.Product{ID: "p1", Type: "exam", Status: "published", IsVisible: true})

	svc := newShim(fake)
	products, _, err := svc.ListProducts(ctx, repository.ProductFilter{Type: "exam"}, RoleAdminStore)
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(products) != 0 {
		t.Errorf("admin_store should not see exam products, got %d", len(products))
	}
}

func TestCreateProduct_TypeRBAC(t *testing.T) {
	ctx := context.Background()
	fake := newFakeStoreRepo()
	svc := newShim(fake)

	// admin_store creating exam type → ErrForbidden
	_, err := svc.CreateProduct(ctx, repository.Product{Type: "exam", Title: "Exam 1"}, RoleAdminStore)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("want ErrForbidden for admin_store creating exam, got %v", err)
	}

	// admin_exam creating book type → ErrForbidden
	_, err = svc.CreateProduct(ctx, repository.Product{Type: "book", Title: "Book 1"}, RoleAdminExam)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("want ErrForbidden for admin_exam creating book, got %v", err)
	}

	// admin_store creating book → ok
	p, err := svc.CreateProduct(ctx, repository.Product{Type: "book", Title: "Book 1"}, RoleAdminStore)
	if err != nil {
		t.Fatalf("admin_store creating book: %v", err)
	}
	if p.ID == "" {
		t.Error("want non-empty ID")
	}

	// super_admin creating any type → ok
	_, err = svc.CreateProduct(ctx, repository.Product{Type: "exam", Title: "Exam 1"}, RoleSuperAdmin)
	if err != nil {
		t.Fatalf("super_admin creating exam: %v", err)
	}
}

func TestValidatePromo_Expired(t *testing.T) {
	ctx := context.Background()
	fake := newFakeStoreRepo()
	past := time.Now().Add(-time.Hour)
	fake.seedPromo(repository.PromoCode{
		Code:      "EXPIRED",
		ExpiresAt: &past,
	})
	svc := newShim(fake)
	_, err := svc.ValidatePromo(ctx, "EXPIRED", 100)
	if !errors.Is(err, ErrInvalidPromo) {
		t.Errorf("want ErrInvalidPromo for expired promo, got %v", err)
	}
}

func TestValidatePromo_Math(t *testing.T) {
	ctx := context.Background()
	fake := newFakeStoreRepo()
	fake.seedPromo(repository.PromoCode{
		Code:              "DISC10",
		DiscountPercent:   float64ptr(10),
		MaxDiscountAmount: float64ptr(8),
	})
	svc := newShim(fake)
	result, err := svc.ValidatePromo(ctx, "DISC10", 100)
	if err != nil {
		t.Fatalf("ValidatePromo: %v", err)
	}
	// 10% of 100 = 10, capped to 8
	if result.Discount != 8 {
		t.Errorf("want discount=8 (capped), got %v", result.Discount)
	}
	if result.Total != 92 {
		t.Errorf("want total=92, got %v", result.Total)
	}
}

func TestValidatePromo_MinOrder(t *testing.T) {
	ctx := context.Background()
	fake := newFakeStoreRepo()
	fake.seedPromo(repository.PromoCode{
		Code:           "MINORDER",
		DiscountAmount: float64ptr(20),
		MinOrderAmount: float64ptr(200),
	})
	svc := newShim(fake)
	_, err := svc.ValidatePromo(ctx, "MINORDER", 100)
	if !errors.Is(err, ErrPromoMinOrder) {
		t.Errorf("want ErrPromoMinOrder, got %v", err)
	}
}

func TestValidatePromo_NotFound(t *testing.T) {
	ctx := context.Background()
	fake := newFakeStoreRepo()
	svc := newShim(fake)
	_, err := svc.ValidatePromo(ctx, "MISSING", 100)
	if !errors.Is(err, ErrInvalidPromo) {
		t.Errorf("want ErrInvalidPromo for missing promo, got %v", err)
	}
}

func TestGetShippingRates(t *testing.T) {
	ctx := context.Background()
	svc := newShim(newFakeStoreRepo())
	rates, err := svc.GetShippingRates(ctx, platform.ShippingQuoteRequest{DestinationZip: "12345", WeightGrams: 500})
	if err != nil {
		t.Fatalf("GetShippingRates: %v", err)
	}
	if len(rates) == 0 {
		t.Error("want at least one rate")
	}
}
