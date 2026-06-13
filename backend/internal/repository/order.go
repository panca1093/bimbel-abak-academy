package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type Order struct {
	ID                 uuid.UUID
	StudentID          uuid.UUID
	Status             string
	Subtotal           float64
	Discount           float64
	ShippingAmount     float64
	Total              float64
	PromoCodeID        *uuid.UUID
	ShippingAddress    []byte
	Courier            string
	TrackingNumber     string
	ShippedAt          *time.Time
	PaymentRef         string
	PaymentExpiresAt   *time.Time
	CancellationReason string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Items              []OrderItem
}

type OrderItem struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	ProductType string
	Title       string
	UnitPrice   float64
	Qty         int
	FulfilledAt *time.Time
	CreatedAt   time.Time
}

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

func (r *Repository) MintCart(ctx context.Context, studentID uuid.UUID) (Order, bool, error) {
	order := Order{}
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
				return Order{}, false, err
			}
			return order, false, nil
		}
		return Order{}, false, err
	}
	return order, true, nil
}

func (r *Repository) GetCartByStudentID(ctx context.Context, studentID uuid.UUID) (Order, error) {
	order := Order{}
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
			return Order{}, nil
		}
		return Order{}, err
	}

	items := []OrderItem{}
	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, product_id, product_type, title, unit_price, qty, fulfilled_at, created_at
		 FROM order_item
		 WHERE order_id = $1
		 ORDER BY created_at`,
		order.ID,
	)
	if err != nil {
		return Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		item := OrderItem{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductType,
			&item.Title, &item.UnitPrice, &item.Qty, &item.FulfilledAt, &item.CreatedAt)
		if err != nil {
			return Order{}, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return Order{}, err
	}

	order.Items = items
	return order, nil
}

func (r *Repository) GetOrderByID(ctx context.Context, id uuid.UUID) (Order, error) {
	order := Order{}
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
			return Order{}, nil
		}
		return Order{}, err
	}

	items := []OrderItem{}
	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, product_id, product_type, title, unit_price, qty, fulfilled_at, created_at
		 FROM order_item
		 WHERE order_id = $1
		 ORDER BY created_at`,
		order.ID,
	)
	if err != nil {
		return Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		item := OrderItem{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductType,
			&item.Title, &item.UnitPrice, &item.Qty, &item.FulfilledAt, &item.CreatedAt)
		if err != nil {
			return Order{}, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return Order{}, err
	}

	order.Items = items
	return order, nil
}

func (r *Repository) ListOrders(ctx context.Context, filter OrderFilter) ([]Order, string, error) {
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

	query += fmt.Sprintf(` ORDER BY id ASC LIMIT $%d`, argNum)
	args = append(args, filter.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	orders := []Order{}
	nextCursor := ""

	for rows.Next() {
		order := Order{}
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

func (r *Repository) AddItem(ctx context.Context, orderID uuid.UUID, item OrderItem) error {
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

	items := []OrderItem{}
	for rows.Next() {
		item := OrderItem{}
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
