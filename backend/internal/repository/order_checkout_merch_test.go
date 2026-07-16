package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// Site f (checkout stock enforcement): merchandise must be range-checked and
// decremented at checkout exactly like a book, else it oversells and never
// decrements.
func TestCheckoutOrder_MerchandiseStockEnforcedAndDecremented(t *testing.T) {
	ctx := context.Background()
	pool := newGradingTestPool(t)
	repo := New(pool)

	// Sufficient stock: checkout decrements stock by qty.
	prodID := seedMerchProductRow(t, pool, 5)
	orderID := seedMerchOrderRow(t, pool, prodID, 2)
	tx, err := repo.BeginTx(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.CheckoutOrder(ctx, tx, orderID))
	require.NoError(t, tx.Commit(ctx))
	require.Equal(t, 3, merchStock(t, pool, prodID), "merchandise stock must decrement by qty")

	// Insufficient stock: checkout fails and stock is unchanged.
	prodID2 := seedMerchProductRow(t, pool, 1)
	orderID2 := seedMerchOrderRow(t, pool, prodID2, 3)
	tx2, err := repo.BeginTx(ctx)
	require.NoError(t, err)
	err = repo.CheckoutOrder(ctx, tx2, orderID2)
	require.ErrorIs(t, err, ErrInsufficientStock)
	_ = tx2.Rollback(ctx)
	require.Equal(t, 1, merchStock(t, pool, prodID2), "stock must be unchanged after an insufficient-stock checkout")
}

func TestCheckoutOrder_MedalStockEnforcedAndDecremented(t *testing.T) {
	ctx := context.Background()
	pool := newGradingTestPool(t)
	repo := New(pool)

	prodID := seedPhysicalProductRow(t, pool, "medal", 5)
	orderID := seedPhysicalOrderRow(t, pool, prodID, "medal", 2)
	tx, err := repo.BeginTx(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.CheckoutOrder(ctx, tx, orderID))
	require.NoError(t, tx.Commit(ctx))
	require.Equal(t, 3, merchStock(t, pool, prodID))
}

func seedMerchProductRow(t *testing.T, pool *pgxpool.Pool, stock int) uuid.UUID {
	return seedPhysicalProductRow(t, pool, "merchandise", stock)
}

func seedPhysicalProductRow(t *testing.T, pool *pgxpool.Pool, productType string, stock int) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO product (type, name, price, stock, status) VALUES ($1, 'Academy Tee', 100, $2, 'published') RETURNING id`,
		productType, stock,
	).Scan(&id))
	return id
}

func seedMerchOrderRow(t *testing.T, pool *pgxpool.Pool, productID uuid.UUID, qty int) uuid.UUID {
	return seedPhysicalOrderRow(t, pool, productID, "merchandise", qty)
}

func seedPhysicalOrderRow(t *testing.T, pool *pgxpool.Pool, productID uuid.UUID, productType string, qty int) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	studentID := insertGradingUser(t, pool, "student", "Merch Buyer")
	var orderID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO orders (student_id, status, subtotal, discount, shipping_cost, total)
		 VALUES ($1, 'cart', 0, 0, 0, 0) RETURNING id`, studentID,
	).Scan(&orderID))
	_, err := pool.Exec(ctx,
		`INSERT INTO order_item (id, order_id, product_id, product_type, name, unit_price, qty, jumlah, weight_grams, created_at)
		 VALUES ($1, $2, $3, $4, 'Academy Tee', 100, $5, $6, 0, now())`,
		uuid.New(), orderID, productID, productType, qty, float64(100*qty),
	)
	require.NoError(t, err)
	return orderID
}

func merchStock(t *testing.T, pool *pgxpool.Pool, productID uuid.UUID) int {
	t.Helper()
	var stock int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT stock FROM product WHERE id = $1`, productID).Scan(&stock))
	return stock
}
