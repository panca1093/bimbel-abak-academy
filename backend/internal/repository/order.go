package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"akademi-bimbel/internal/model"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type OrderFilter struct {
	StudentID   *uuid.UUID
	Status      string
	ProductType string
	Cursor      string
	Limit       int
}

type OrderPatch struct {
	ShippingAddress []byte
	Courier         string
	PromoCodeID     *uuid.UUID
	Discount        float64
	ShippingAmount  float64
	Total           float64
}

func (r *Repository) MintCart(ctx context.Context, studentID uuid.UUID) (model.Order, bool, error) {
	order := model.Order{}
	// Try to INSERT; ON CONFLICT DO NOTHING returns nothing if conflict
	err := r.pool.QueryRow(ctx,
		`INSERT INTO orders (student_id, status, subtotal, discount, shipping_amount, total)
		 VALUES ($1, 'cart', 0, 0, 0, 0)
		 ON CONFLICT (student_id) WHERE status = 'cart' DO NOTHING
		 RETURNING id, student_id, status, subtotal, discount, shipping_amount, total,
		           promo_code_id, shipping_address, courier, tracking_number, shipped_at,
		           payment_ref, payment_expires_at, cancellation_reason, created_at, updated_at`,
		studentID,
	).Scan(
		&order.ID, &order.StudentID, &order.Status, &order.Subtotal, &order.Discount,
		&order.ShippingAmount, &order.Total, &order.PromoCodeID, &order.ShippingAddress,
		&order.Courier, &order.TrackingNumber, &order.ShippedAt,
		&order.PaymentRef, &order.PaymentExpiresAt, &order.CancellationReason,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			// INSERT hit conflict; fetch existing cart
			err = r.pool.QueryRow(ctx,
				`SELECT id, student_id, status, subtotal, discount, shipping_amount, total,
				        promo_code_id, shipping_address, courier, tracking_number, shipped_at,
				        payment_ref, payment_expires_at, cancellation_reason, created_at, updated_at
				 FROM orders
				 WHERE student_id = $1 AND status = 'cart'`,
				studentID,
			).Scan(
				&order.ID, &order.StudentID, &order.Status, &order.Subtotal, &order.Discount,
				&order.ShippingAmount, &order.Total, &order.PromoCodeID, &order.ShippingAddress,
				&order.Courier, &order.TrackingNumber, &order.ShippedAt,
				&order.PaymentRef, &order.PaymentExpiresAt, &order.CancellationReason,
				&order.CreatedAt, &order.UpdatedAt,
			)
			if err != nil {
				return model.Order{}, false, err
			}
			return order, false, nil
		}
		return model.Order{}, false, err
	}
	return order, true, nil
}

func (r *Repository) GetCartByStudentID(ctx context.Context, studentID uuid.UUID) (model.Order, error) {
	order := model.Order{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, status, subtotal, discount, shipping_amount, total,
		        promo_code_id, shipping_address, courier, tracking_number, shipped_at,
		        payment_ref, payment_expires_at, cancellation_reason, created_at, updated_at
		 FROM orders
		 WHERE student_id = $1 AND status = 'cart'`,
		studentID,
	).Scan(
		&order.ID, &order.StudentID, &order.Status, &order.Subtotal, &order.Discount,
		&order.ShippingAmount, &order.Total, &order.PromoCodeID, &order.ShippingAddress,
		&order.Courier, &order.TrackingNumber, &order.ShippedAt,
		&order.PaymentRef, &order.PaymentExpiresAt, &order.CancellationReason,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return model.Order{}, nil
		}
		return model.Order{}, err
	}

	items := []model.OrderItem{}
	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, product_id, product_type, title, unit_price, qty, fulfilled_at, created_at
		 FROM order_item
		 WHERE order_id = $1
		 ORDER BY created_at`,
		order.ID,
	)
	if err != nil {
		return model.Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		item := model.OrderItem{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductType,
			&item.Title, &item.UnitPrice, &item.Qty, &item.FulfilledAt, &item.CreatedAt)
		if err != nil {
			return model.Order{}, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return model.Order{}, err
	}

	order.Items = items
	return order, nil
}

func (r *Repository) GetOrderByID(ctx context.Context, id uuid.UUID) (model.Order, error) {
	order := model.Order{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, status, subtotal, discount, shipping_amount, total,
		        promo_code_id, shipping_address, courier, tracking_number, shipped_at,
		        payment_ref, payment_expires_at, cancellation_reason, created_at, updated_at
		 FROM orders
		 WHERE id = $1`,
		id,
	).Scan(
		&order.ID, &order.StudentID, &order.Status, &order.Subtotal, &order.Discount,
		&order.ShippingAmount, &order.Total, &order.PromoCodeID, &order.ShippingAddress,
		&order.Courier, &order.TrackingNumber, &order.ShippedAt,
		&order.PaymentRef, &order.PaymentExpiresAt, &order.CancellationReason,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return model.Order{}, nil
		}
		return model.Order{}, err
	}

	items := []model.OrderItem{}
	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, product_id, product_type, title, unit_price, qty, fulfilled_at, created_at
		 FROM order_item
		 WHERE order_id = $1
		 ORDER BY created_at`,
		order.ID,
	)
	if err != nil {
		return model.Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		item := model.OrderItem{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductType,
			&item.Title, &item.UnitPrice, &item.Qty, &item.FulfilledAt, &item.CreatedAt)
		if err != nil {
			return model.Order{}, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return model.Order{}, err
	}

	order.Items = items
	return order, nil
}

func (r *Repository) ListOrders(ctx context.Context, filter OrderFilter) ([]model.Order, string, error) {
	if filter.Limit == 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	query := `SELECT id, student_id, status, subtotal, discount, shipping_amount, total,
	                 promo_code_id, shipping_address, courier, tracking_number, shipped_at,
	                 payment_ref, payment_expires_at, cancellation_reason, created_at, updated_at
	          FROM orders
	          WHERE 1=1`

	args := []interface{}{}
	argNum := 1

	if filter.StudentID != nil {
		query += fmt.Sprintf(` AND student_id = $%d`, argNum)
		args = append(args, *filter.StudentID)
		argNum++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(` AND status = $%d`, argNum)
		args = append(args, filter.Status)
		argNum++
	}

	if filter.ProductType != "" {
		query += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM order_item WHERE order_item.order_id = orders.id AND product_type = $%d)`, argNum)
		args = append(args, filter.ProductType)
		argNum++
	}

	if filter.Cursor != "" {
		query += fmt.Sprintf(` AND id > $%d`, argNum)
		args = append(args, filter.Cursor)
		argNum++
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, argNum)
	args = append(args, filter.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	orders := []model.Order{}
	nextCursor := ""

	for rows.Next() {
		order := model.Order{}
		err := rows.Scan(
			&order.ID, &order.StudentID, &order.Status, &order.Subtotal, &order.Discount,
			&order.ShippingAmount, &order.Total, &order.PromoCodeID, &order.ShippingAddress,
			&order.Courier, &order.TrackingNumber, &order.ShippedAt,
			&order.PaymentRef, &order.PaymentExpiresAt, &order.CancellationReason,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, "", err
		}

		if len(orders) < filter.Limit {
			orders = append(orders, order)
		} else {
			nextCursor = order.ID.String()
		}
	}

	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	return orders, nextCursor, nil
}

func (r *Repository) AddItem(ctx context.Context, orderID uuid.UUID, item model.OrderItem) error {
	item.ID = uuid.New()
	item.OrderID = orderID
	_, err := r.pool.Exec(ctx,
		`INSERT INTO order_item (id, order_id, product_id, product_type, title, unit_price, qty, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, now())`,
		item.ID, item.OrderID, item.ProductID, item.ProductType, item.Title, item.UnitPrice, item.Qty,
	)
	return err
}

func (r *Repository) RemoveItem(ctx context.Context, orderID, itemID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM order_item WHERE id = $1 AND order_id = $2`,
		itemID, orderID,
	)
	return err
}

func (r *Repository) PatchCart(ctx context.Context, orderID uuid.UUID, patch OrderPatch) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders
		 SET shipping_address = $1, courier = $2, promo_code_id = $3,
		     discount = $4, shipping_amount = $5, total = $6, updated_at = now()
		 WHERE id = $7`,
		patch.ShippingAddress, patch.Courier, patch.PromoCodeID,
		patch.Discount, patch.ShippingAmount, patch.Total, orderID,
	)
	return err
}

func (r *Repository) SetOrderStatus(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, status, reason string) error {
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx,
			`UPDATE orders
			 SET status = $1, cancellation_reason = $2, updated_at = now()
			 WHERE id = $3`,
			status, reason, orderID,
		)
	} else {
		_, err = r.pool.Exec(ctx,
			`UPDATE orders
			 SET status = $1, cancellation_reason = $2, updated_at = now()
			 WHERE id = $3`,
			status, reason, orderID,
		)
	}
	return err
}

func (r *Repository) SetShipped(ctx context.Context, orderID uuid.UUID, trackingNumber string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders
		 SET tracking_number = $1, shipped_at = now(), updated_at = now()
		 WHERE id = $2`,
		trackingNumber, orderID,
	)
	return err
}

func (r *Repository) SetPaymentRef(ctx context.Context, orderID uuid.UUID, ref string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders
		 SET payment_ref = $1, payment_expires_at = $2, updated_at = now()
		 WHERE id = $3`,
		ref, expiresAt, orderID,
	)
	return err
}

func (r *Repository) GetExpiredPaymentOrders(ctx context.Context, limit int) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id
		FROM orders
		WHERE status = 'payment_pending'
		  AND payment_expires_at < now()
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		orderIDs = append(orderIDs, id)
	}
	return orderIDs, rows.Err()
}

func (r *Repository) CheckoutOrder(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error {
	// Fetch order items
	rows, err := tx.Query(ctx,
		`SELECT id, order_id, product_id, product_type, title, unit_price, qty, fulfilled_at, created_at
		 FROM order_item
		 WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	items := []model.OrderItem{}
	for rows.Next() {
		item := model.OrderItem{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductType,
			&item.Title, &item.UnitPrice, &item.Qty, &item.FulfilledAt, &item.CreatedAt)
		if err != nil {
			return err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// For each item, SELECT product FOR UPDATE and decrement stock
	for _, item := range items {
		var currentStock int
		err := tx.QueryRow(ctx,
			`SELECT stock FROM product WHERE id = $1 FOR UPDATE`,
			item.ProductID,
		).Scan(&currentStock)
		if err != nil {
			return err
		}

		if currentStock < item.Qty {
			return ErrInsufficientStock
		}
		newStock := currentStock - item.Qty

		_, err = tx.Exec(ctx,
			`UPDATE product SET stock = $1, updated_at = now() WHERE id = $2`,
			newStock, item.ProductID,
		)
		if err != nil {
			return err
		}
	}

	// Set order status to payment_pending
	_, err = tx.Exec(ctx,
		`UPDATE orders
		 SET status = 'payment_pending', updated_at = now()
		 WHERE id = $1`,
		orderID,
	)
	return err
}

func (r *Repository) InsertWebhookLog(ctx context.Context, tx pgx.Tx, eventType string, payload []byte, paymentRef string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO webhook_log (event_type, payload, payment_ref)
		 VALUES ($1, $2, $3)`,
		eventType, payload, paymentRef,
	)
	return err
}

func (r *Repository) InsertAuditLog(ctx context.Context, tx pgx.Tx, actorID, targetType, targetID, action string) error {
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx,
			`INSERT INTO audit_log (actor_id, target_type, target_id, action)
			 VALUES ($1, $2, $3, $4)`,
			actorID, targetType, targetID, action,
		)
	} else {
		_, err = r.pool.Exec(ctx,
			`INSERT INTO audit_log (actor_id, target_type, target_id, action)
			 VALUES ($1, $2, $3, $4)`,
			actorID, targetType, targetID, action,
		)
	}
	return err
}

func (r *Repository) ClearOrderTracking(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error {
	_, err := tx.Exec(ctx,
		`UPDATE orders
		 SET tracking_number = '', shipped_at = NULL, updated_at = now()
		 WHERE id = $1`,
		orderID,
	)
	return err
}

func (r *Repository) GetRevenue(ctx context.Context, from, to time.Time) (map[string]interface{}, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT COALESCE(SUM(o.total), 0) as total, oi.product_type, COUNT(*) as count
		FROM orders o
		JOIN order_item oi ON o.id = oi.order_id
		WHERE o.status IN ('paid', 'processing', 'shipped')
		  AND o.created_at BETWEEN $1 AND $2
		GROUP BY oi.product_type
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]interface{}{
		"total":   0.0,
		"by_type": map[string]interface{}{},
	}
	var grandTotal float64
	byType := map[string]interface{}{}

	for rows.Next() {
		var total float64
		var productType string
		var count int
		if err := rows.Scan(&total, &productType, &count); err != nil {
			return nil, err
		}
		grandTotal += total
		byType[productType] = map[string]interface{}{
			"total": total,
			"count": count,
		}
	}

	result["total"] = grandTotal
	result["by_type"] = byType
	return result, rows.Err()
}
