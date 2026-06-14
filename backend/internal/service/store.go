package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/repository"
)

var (
	ErrForbidden         = errors.New("forbidden")
	ErrProductNotFound   = errors.New("product not found")
	ErrInvalidPromo      = errors.New("invalid or expired promo code")
	ErrPromoMinOrder     = errors.New("order subtotal below promo minimum")
	ErrOutOfStock        = errors.New("product out of stock")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrOrderNotEditable  = errors.New("order not editable")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidSignature  = errors.New("invalid signature")
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

func (s *Service) MintCart(ctx context.Context, studentID string) (repository.Order, bool, error) {
	id, err := parseUUID(studentID)
	if err != nil {
		return repository.Order{}, false, err
	}
	return s.storeRepo.MintCart(ctx, id)
}

func (s *Service) AddItem(ctx context.Context, studentID, orderID, productID string, qty int) error {
	oID, err := parseUUID(orderID)
	if err != nil {
		return err
	}
	sID, err := parseUUID(studentID)
	if err != nil {
		return err
	}
	pID, err := parseUUID(productID)
	if err != nil {
		return err
	}

	order, err := s.storeRepo.GetOrderByID(ctx, oID)
	if err != nil {
		return err
	}
	if order.ID.String() == "" {
		return ErrOrderNotFound
	}
	if order.StudentID != sID {
		return ErrOrderNotFound
	}
	if order.Status != "cart" {
		return ErrOrderNotEditable
	}

	product, err := s.storeRepo.GetProductByID(ctx, pID.String())
	if err != nil {
		return err
	}
	if product == nil {
		return ErrProductNotFound
	}
	if product.Stock == 0 {
		return ErrOutOfStock
	}

	item := repository.OrderItem{
		ProductID:   pID,
		ProductType: product.Type,
		Title:       product.Title,
		UnitPrice:   float64(product.Price) / 100,
		Qty:         qty,
	}
	return s.storeRepo.AddItem(ctx, oID, item)
}

func (s *Service) RemoveItem(ctx context.Context, studentID, orderID, itemID string) error {
	oID, err := parseUUID(orderID)
	if err != nil {
		return err
	}
	sID, err := parseUUID(studentID)
	if err != nil {
		return err
	}
	iID, err := parseUUID(itemID)
	if err != nil {
		return err
	}

	order, err := s.storeRepo.GetOrderByID(ctx, oID)
	if err != nil {
		return err
	}
	if order.ID.String() == "" {
		return ErrOrderNotFound
	}
	if order.StudentID != sID {
		return ErrOrderNotFound
	}

	return s.storeRepo.RemoveItem(ctx, oID, iID)
}

type CartPatch struct {
	ShippingAddress []byte
	Courier         string
	PromoCode       *string
}

func (s *Service) PatchCart(ctx context.Context, studentID, orderID string, patch CartPatch) error {
	oID, err := parseUUID(orderID)
	if err != nil {
		return err
	}
	sID, err := parseUUID(studentID)
	if err != nil {
		return err
	}

	order, err := s.storeRepo.GetOrderByID(ctx, oID)
	if err != nil {
		return err
	}
	if order.ID.String() == "" {
		return ErrOrderNotFound
	}
	if order.StudentID != sID {
		return ErrOrderNotFound
	}
	if order.Status != "cart" {
		return ErrOrderNotEditable
	}

	repoPatch := repository.OrderPatch{
		ShippingAddress: patch.ShippingAddress,
		Courier:         patch.Courier,
		Discount:        order.Discount,
		ShippingAmount:  order.ShippingAmount,
		Total:           order.Total,
	}

	if patch.PromoCode != nil && *patch.PromoCode != "" {
		validation, err := s.ValidatePromo(ctx, *patch.PromoCode, order.Subtotal)
		if err != nil {
			return err
		}
		repoPatch.Discount = validation.Discount
		repoPatch.Total = validation.Total
	}

	return s.storeRepo.PatchCart(ctx, oID, repoPatch)
}

type CheckoutResult struct {
	PaymentRef      string
	PaymentExpiresAt time.Time
}

func (s *Service) Checkout(ctx context.Context, studentID, orderID, key string) (CheckoutResult, error) {
	oID, err := parseUUID(orderID)
	if err != nil {
		return CheckoutResult{}, err
	}
	sID, err := parseUUID(studentID)
	if err != nil {
		return CheckoutResult{}, err
	}

	cacheKey := "idempotency:checkout:" + key
	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		return CheckoutResult{PaymentRef: cached}, nil
	}

	order, err := s.storeRepo.GetOrderByID(ctx, oID)
	if err != nil {
		return CheckoutResult{}, err
	}
	if order.ID.String() == "" {
		return CheckoutResult{}, ErrOrderNotFound
	}
	if order.StudentID != sID {
		return CheckoutResult{}, ErrOrderNotFound
	}
	if order.Status != "cart" {
		return CheckoutResult{}, ErrOrderNotEditable
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return CheckoutResult{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.CheckoutOrder(ctx, tx, oID); err != nil {
		if errors.Is(err, repository.ErrInsufficientStock) {
			return CheckoutResult{}, ErrInsufficientStock
		}
		return CheckoutResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return CheckoutResult{}, err
	}

	paymentRef := "pay_" + oID.String()[:8]
	expiresAt := time.Now().Add(24 * time.Hour)

	if err := s.storeRepo.SetPaymentRef(ctx, oID, paymentRef, expiresAt); err != nil {
		return CheckoutResult{}, err
	}

	result := CheckoutResult{
		PaymentRef:       paymentRef,
		PaymentExpiresAt: expiresAt,
	}

	if err := s.rdb.Set(ctx, cacheKey, paymentRef, 24*time.Hour).Err(); err != nil {
		return CheckoutResult{}, err
	}

	return result, nil
}

func (s *Service) RetryPayment(ctx context.Context, studentID, orderID, key string) (CheckoutResult, error) {
	oID, err := parseUUID(orderID)
	if err != nil {
		return CheckoutResult{}, err
	}
	sID, err := parseUUID(studentID)
	if err != nil {
		return CheckoutResult{}, err
	}

	cacheKey := "idempotency:retry:" + key
	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		return CheckoutResult{PaymentRef: cached}, nil
	}

	order, err := s.storeRepo.GetOrderByID(ctx, oID)
	if err != nil {
		return CheckoutResult{}, err
	}
	if order.ID.String() == "" {
		return CheckoutResult{}, ErrOrderNotFound
	}
	if order.StudentID != sID {
		return CheckoutResult{}, ErrOrderNotFound
	}
	if order.Status != "payment_expired" && order.Status != "payment_failed" {
		return CheckoutResult{}, ErrOrderNotEditable
	}

	paymentRef := "pay_" + oID.String()[:8]
	expiresAt := time.Now().Add(24 * time.Hour)

	if err := s.storeRepo.SetPaymentRef(ctx, oID, paymentRef, expiresAt); err != nil {
		return CheckoutResult{}, err
	}

	result := CheckoutResult{
		PaymentRef:       paymentRef,
		PaymentExpiresAt: expiresAt,
	}

	if err := s.rdb.Set(ctx, cacheKey, paymentRef, 24*time.Hour).Err(); err != nil {
		return CheckoutResult{}, err
	}

	return result, nil
}

func (s *Service) ListStudentOrders(ctx context.Context, studentID string, cursor string, limit int) ([]repository.Order, string, error) {
	sID, err := parseUUID(studentID)
	if err != nil {
		return nil, "", err
	}
	orders, nextCursor, err := s.storeRepo.ListOrders(ctx, repository.OrderFilter{
		StudentID: &sID,
		Cursor:    cursor,
		Limit:     limit,
	})
	if err != nil {
		return nil, "", err
	}

	var filtered []repository.Order
	for _, o := range orders {
		if o.Status != "cart" {
			filtered = append(filtered, o)
		}
	}
	return filtered, nextCursor, nil
}

func (s *Service) GetStudentOrder(ctx context.Context, studentID, orderID string) (repository.Order, error) {
	oID, err := parseUUID(orderID)
	if err != nil {
		return repository.Order{}, err
	}
	sID, err := parseUUID(studentID)
	if err != nil {
		return repository.Order{}, err
	}

	order, err := s.storeRepo.GetOrderByID(ctx, oID)
	if err != nil {
		return repository.Order{}, err
	}
	if order.ID.String() == "" {
		return repository.Order{}, ErrOrderNotFound
	}
	if order.StudentID != sID {
		return repository.Order{}, ErrOrderNotFound
	}
	return order, nil
}

func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

// Admin order methods

func (s *Service) AdminListOrders(ctx context.Context, filter repository.OrderFilter) ([]repository.Order, string, error) {
	return s.storeRepo.ListOrders(ctx, filter)
}

func (s *Service) AdminGetOrder(ctx context.Context, orderID string) (repository.Order, error) {
	id, err := parseUUID(orderID)
	if err != nil {
		return repository.Order{}, err
	}
	return s.storeRepo.GetOrderByID(ctx, id)
}

func (s *Service) AdminConfirmOrder(ctx context.Context, orderID, key string) error {
	id, err := parseUUID(orderID)
	if err != nil {
		return err
	}

	cacheKey := "idempotency:confirm:" + key
	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		return nil
	}

	order, err := s.storeRepo.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}
	if order.ID.String() == "" {
		return ErrOrderNotFound
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.SetOrderStatus(ctx, tx, id, "paid", ""); err != nil {
		return err
	}

	if err := s.storeRepo.InsertOutboxEvent(ctx, tx, id, "OrderPaid", nil); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if err := s.rdb.Set(ctx, cacheKey, "ok", 24*time.Hour).Err(); err != nil {
		return err
	}

	return nil
}

func (s *Service) AdminShipOrder(ctx context.Context, orderID, trackingNumber string) error {
	id, err := parseUUID(orderID)
	if err != nil {
		return err
	}

	order, err := s.storeRepo.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}
	if order.ID.String() == "" {
		return ErrOrderNotFound
	}

	if order.Status != "paid" && order.Status != "processing" {
		return errors.New("order not in shippable status")
	}

	return s.storeRepo.SetShipped(ctx, id, trackingNumber)
}

func (s *Service) AdminRefundOrder(ctx context.Context, orderID string) error {
	id, err := parseUUID(orderID)
	if err != nil {
		return err
	}

	order, err := s.storeRepo.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}
	if order.ID.String() == "" {
		return ErrOrderNotFound
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.SetOrderStatus(ctx, tx, id, "cancelled", "refunded"); err != nil {
		return err
	}

	if err := s.storeRepo.RevokeEnrollmentsByOrder(ctx, tx, id); err != nil {
		return err
	}

	if err := s.storeRepo.ExpireExamRegistrationsByOrder(ctx, tx, id); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) AdminReconcileOrder(ctx context.Context, orderID, key string) error {
	id, err := parseUUID(orderID)
	if err != nil {
		return err
	}

	cacheKey := "idempotency:reconcile:" + key
	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		return nil
	}

	status, err := s.payment.QueryStatus(ctx, "")
	if err != nil {
		return err
	}

	if status.Paid {
		if err := s.storeRepo.SetOrderStatus(ctx, nil, id, "paid", ""); err != nil {
			return err
		}
	}

	if err := s.rdb.Set(ctx, cacheKey, "ok", 24*time.Hour).Err(); err != nil {
		return err
	}

	return nil
}

// Admin promo methods

func (s *Service) AdminListPromoCodes(ctx context.Context) ([]repository.PromoCode, error) {
	return s.storeRepo.ListPromoCodes(ctx)
}

func (s *Service) AdminCreatePromoCode(ctx context.Context, p repository.PromoCode) (repository.PromoCode, error) {
	return s.storeRepo.CreatePromoCode(ctx, p)
}

func (s *Service) AdminUpdatePromoCode(ctx context.Context, id string, maxUses *int, expiresAt *time.Time) error {
	pID, err := parseUUID(id)
	if err != nil {
		return err
	}
	return s.storeRepo.UpdatePromoCode(ctx, pID, maxUses, expiresAt)
}

func (s *Service) AdminDeletePromoCode(ctx context.Context, id string) error {
	pID, err := parseUUID(id)
	if err != nil {
		return err
	}
	return s.storeRepo.DeletePromoCode(ctx, pID)
}

// Admin revenue method

func (s *Service) AdminGetRevenue(ctx context.Context, from, to time.Time) (map[string]interface{}, error) {
	// Placeholder - will aggregate orders with paid/processing/shipped status
	return map[string]interface{}{
		"total": 0.0,
	}, nil
}

// Payment webhook handler

func (s *Service) HandlePaymentWebhook(ctx context.Context, payload []byte, signature, key string) error {
	if !s.payment.VerifySignature(payload, signature) {
		return ErrInvalidSignature
	}

	cacheKey := "idempotency:webhook:" + key
	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		return nil
	}

	// Placeholder - will parse webhook, update order status, insert outbox event
	if err := s.rdb.Set(ctx, cacheKey, "ok", 24*time.Hour).Err(); err != nil {
		return err
	}

	return nil
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
