package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

var (
	ErrForbidden         = errors.New("forbidden")
	ErrProductNotFound   = errors.New("product not found")
	ErrCourseNotFound    = errors.New("course not found")
	ErrInvalidPromo      = errors.New("invalid or expired promo code")
	ErrPromoMinOrder     = errors.New("order subtotal below promo minimum")
	ErrOutOfStock        = errors.New("product out of stock")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrOrderNotEditable  = errors.New("order not editable")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrCourseLinkRequired = errors.New("course product requires at least one linked course")
)

type PromoValidation struct {
	Code     string
	Discount float64
	Total    float64
}

func (s *Service) ListProducts(ctx context.Context, filter repository.ProductFilter, role string) ([]model.Product, string, error) {
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
		filter.VisibleOnly = true
		filter.Status = "published"
	}
	return s.storeRepo.ListProducts(ctx, filter)
}

func (s *Service) GetProduct(ctx context.Context, id string, role string) (model.Product, error) {
	p, err := s.storeRepo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Product{}, ErrProductNotFound
		}
		return model.Product{}, err
	}
	if role == RoleStudent || role == "" {
		if p.Status != "published" {
			return model.Product{}, ErrProductNotFound
		}
	}
	if p.Type == "course" {
		pID, err := parseUUID(p.ID)
		if err == nil {
			courses, err := s.storeRepo.GetCoursesByProductID(ctx, pID)
			if err == nil {
				for _, c := range courses {
					p.CourseIDs = append(p.CourseIDs, c.ID.String())
				}
			}
		}
	}
	return *p, nil
}

func (s *Service) CreateProduct(ctx context.Context, p model.Product, role string) (model.Product, error) {
	if err := checkTypeRBAC(role, p.Type); err != nil {
		return model.Product{}, err
	}
	if err := s.storeRepo.CreateProduct(ctx, &p); err != nil {
		return model.Product{}, err
	}
	return p, nil
}

func (s *Service) CreateProductWithCourses(ctx context.Context, p model.Product, courseIDs []string, role string) (model.Product, error) {
	if err := checkTypeRBAC(role, p.Type); err != nil {
		return model.Product{}, err
	}

	if p.Type == "course" && len(courseIDs) < 1 {
		return model.Product{}, ErrCourseLinkRequired
	}

	var ids []uuid.UUID
	for _, cid := range courseIDs {
		parsed, err := parseUUID(cid)
		if err != nil {
			return model.Product{}, err
		}
		ids = append(ids, parsed)
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return model.Product{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.CreateProductWithCourses(ctx, tx, &p, ids); err != nil {
		return model.Product{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Product{}, err
	}

	return p, nil
}

func (s *Service) UpdateProductWithCourses(ctx context.Context, id string, p model.Product, courseIDs []string, role string) (model.Product, error) {
	existing, err := s.storeRepo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Product{}, ErrProductNotFound
		}
		return model.Product{}, err
	}
	if err := checkTypeRBAC(role, existing.Type); err != nil {
		return model.Product{}, err
	}

	var ids []uuid.UUID
	for _, cid := range courseIDs {
		parsed, err := parseUUID(cid)
		if err != nil {
			return model.Product{}, err
		}
		ids = append(ids, parsed)
	}

	pID, err := parseUUID(id)
	if err != nil {
		return model.Product{}, err
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return model.Product{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.UpdateProductTx(ctx, tx, id, &p); err != nil {
		return model.Product{}, err
	}
	if err := s.storeRepo.ReplaceProductCourses(ctx, tx, pID, ids); err != nil {
		return model.Product{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Product{}, err
	}

	p.ID = id
	p.CourseIDs = courseIDs
	return p, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id string, p model.Product, role string) (model.Product, error) {
	existing, err := s.storeRepo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Product{}, ErrProductNotFound
		}
		return model.Product{}, err
	}
	if err := checkTypeRBAC(role, existing.Type); err != nil {
		return model.Product{}, err
	}
	if err := s.storeRepo.UpdateProduct(ctx, id, &p); err != nil {
		return model.Product{}, err
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
	if promo.MaxUses != nil && promo.UsedCount >= *promo.MaxUses {
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

func (s *Service) GetShippingRates(ctx context.Context, req ShippingQuoteRequest) ([]CourierRate, error) {
	return s.logistics.GetRates(ctx, req)
}

func (s *Service) MintCart(ctx context.Context, studentID string) (model.Order, bool, error) {
	id, err := parseUUID(studentID)
	if err != nil {
		return model.Order{}, false, err
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
	if product.Type == "book" && product.Stock == 0 {
		return ErrOutOfStock
	}

	item := model.OrderItem{
		ProductID:   pID,
		ProductType: product.Type,
		Name:        product.Name,
		UnitPrice:   float64(product.Price) / 100,
		Qty:         qty,
		WeightGrams: product.WeightGrams,
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
		SelectedCourier: patch.Courier,
		Discount:        order.Discount,
		ShippingCost:    order.ShippingCost,
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
	GatewayRef       string
	SnapToken        string
	PaymentExpiresAt time.Time
}

type OrderPaidPayload struct {
	OrderID string                 `json:"order_id"`
	Items   []OrderPaidPayloadItem `json:"items"`
}

type OrderPaidPayloadItem struct {
	ProductID   string `json:"product_id"`
	ProductType string `json:"product_type"`
	Qty         int    `json:"qty"`
}

type MidtransNotification struct {
	TransactionStatus string `json:"transaction_status"`
	OrderID           string `json:"order_id"`
	TransactionID     string `json:"transaction_id"`
	GrossAmount       string `json:"gross_amount"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
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
		return CheckoutResult{GatewayRef: cached}, nil
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

	paymentResp, err := s.payment.CreatePayment(ctx, PaymentRequest{
		OrderID:   oID.String(),
		Amount:    int64(order.Total),
		ExpiresIn: 24 * time.Hour,
	})
	if err != nil {
		return CheckoutResult{}, err
	}

	if err := s.storeRepo.SetPaymentRef(ctx, oID, paymentResp.GatewayRef, paymentResp.ExpiresAt); err != nil {
		return CheckoutResult{}, err
	}

	result := CheckoutResult{
		GatewayRef:       paymentResp.GatewayRef,
		SnapToken:        paymentResp.SnapToken,
		PaymentExpiresAt: paymentResp.ExpiresAt,
	}

	if err := s.rdb.Set(ctx, cacheKey, paymentResp.GatewayRef, 24*time.Hour).Err(); err != nil {
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
		return CheckoutResult{GatewayRef: cached}, nil
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
	if order.Status != "payment_expired" {
		return CheckoutResult{}, ErrOrderNotEditable
	}

	paymentResp, err := s.payment.CreatePayment(ctx, PaymentRequest{
		OrderID:   oID.String(),
		Amount:    int64(order.Total),
		ExpiresIn: 24 * time.Hour,
	})
	if err != nil {
		return CheckoutResult{}, err
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return CheckoutResult{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.SetPaymentRef(ctx, oID, paymentResp.GatewayRef, paymentResp.ExpiresAt); err != nil {
		return CheckoutResult{}, err
	}

	if err := s.storeRepo.SetOrderStatus(ctx, tx, oID, "payment_pending", ""); err != nil {
		return CheckoutResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return CheckoutResult{}, err
	}

	result := CheckoutResult{
		GatewayRef:       paymentResp.GatewayRef,
		SnapToken:        paymentResp.SnapToken,
		PaymentExpiresAt: paymentResp.ExpiresAt,
	}

	if err := s.rdb.Set(ctx, cacheKey, paymentResp.GatewayRef, 24*time.Hour).Err(); err != nil {
		return CheckoutResult{}, err
	}

	return result, nil
}

func (s *Service) ListStudentOrders(ctx context.Context, studentID string, cursor string, limit int) ([]model.Order, string, error) {
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

	var filtered []model.Order
	for _, o := range orders {
		if o.Status != "cart" {
			filtered = append(filtered, o)
		}
	}
	return filtered, nextCursor, nil
}

func (s *Service) GetStudentOrder(ctx context.Context, studentID, orderID string) (model.Order, error) {
	oID, err := parseUUID(orderID)
	if err != nil {
		return model.Order{}, err
	}
	sID, err := parseUUID(studentID)
	if err != nil {
		return model.Order{}, err
	}

	order, err := s.storeRepo.GetOrderByID(ctx, oID)
	if err != nil {
		return model.Order{}, err
	}
	if order.ID.String() == "" {
		return model.Order{}, ErrOrderNotFound
	}
	if order.StudentID != sID {
		return model.Order{}, ErrOrderNotFound
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

func (s *Service) AdminListOrders(ctx context.Context, filter repository.OrderFilter) ([]model.Order, string, error) {
	return s.storeRepo.ListOrders(ctx, filter)
}

func (s *Service) AdminGetOrder(ctx context.Context, orderID string) (model.Order, error) {
	id, err := parseUUID(orderID)
	if err != nil {
		return model.Order{}, err
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

	payload := OrderPaidPayload{OrderID: id.String()}
	for _, item := range order.Items {
		payload.Items = append(payload.Items, OrderPaidPayloadItem{
			ProductID:   item.ProductID.String(),
			ProductType: item.ProductType,
			Qty:         item.Qty,
		})
	}

	if err := s.storeRepo.InsertOutboxEvent(ctx, tx, id, "OrderPaid", payload); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if err := s.rdb.Set(ctx, cacheKey, "ok", 24*time.Hour).Err(); err != nil {
		return err
	}

	// Push notification (best-effort; non-fatal error)
	student, _ := s.storeRepo.GetUserByID(ctx, order.StudentID.String())
	studentName := "Student"
	if student != nil {
		studentName = student.Name
	}
	notif := PurchaseNotification{
		ID:          uuid.New().String(),
		Type:        "order_confirmed",
		OrderID:     order.ID,
		StudentName: studentName,
		Amount:      int64(order.Total * 100),
		CreatedAt:   time.Now(),
		Read:        false,
	}
	_ = s.PushPurchaseNotification(ctx, RoleAdminStore, notif)

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
	if err := s.storeRepo.ClearOrderTracking(ctx, tx, id); err != nil {
		return err
	}

	if err := s.storeRepo.InsertAuditLog(ctx, tx, "", "order", id.String(), "refund"); err != nil {
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

func (s *Service) AdminListPromoCodes(ctx context.Context) ([]model.PromoCode, error) {
	return s.storeRepo.ListPromoCodes(ctx)
}

func (s *Service) AdminCreatePromoCode(ctx context.Context, p model.PromoCode) (model.PromoCode, error) {
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
	return s.storeRepo.GetRevenue(ctx, from, to)
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

	var notif MidtransNotification
	if err := json.Unmarshal(payload, &notif); err != nil {
		return err
	}

	switch notif.TransactionStatus {
	case "settlement", "capture":
		// existing paid flow
	default:
		slog.Info("midtrans webhook ignored", "transaction_status", notif.TransactionStatus, "order_id", notif.OrderID)
		return nil
	}

	orderID, err := parseUUID(notif.OrderID)
	if err != nil {
		return err
	}

	order, err := s.storeRepo.GetOrderByID(ctx, orderID)
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

	if err := s.storeRepo.InsertWebhookLog(ctx, tx, "payment_success", payload, notif.OrderID); err != nil {
		return err
	}

	if err := s.storeRepo.SetOrderStatus(ctx, tx, orderID, "paid", ""); err != nil {
		return err
	}

	outboxPayload := OrderPaidPayload{OrderID: orderID.String()}
	for _, item := range order.Items {
		outboxPayload.Items = append(outboxPayload.Items, OrderPaidPayloadItem{
			ProductID:   item.ProductID.String(),
			ProductType: item.ProductType,
			Qty:         item.Qty,
		})
	}

	if err := s.storeRepo.InsertOutboxEvent(ctx, tx, orderID, "OrderPaid", outboxPayload); err != nil {
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
