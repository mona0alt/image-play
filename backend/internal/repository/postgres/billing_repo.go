package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"image-play/internal/domain/billing"
)

type BillingRepo struct {
	db *sql.DB
}

func NewBillingRepo(db *sql.DB) *BillingRepo {
	return &BillingRepo{db: db}
}

func (r *BillingRepo) GetUser(ctx context.Context, id int64) (*billing.User, error) {
	query := `SELECT id, balance, free_quota FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var u billing.User
	err := row.Scan(&u.ID, &u.Balance, &u.FreeQuota)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *BillingRepo) CreateUser(ctx context.Context, id int64, balance float64) error {
	query := `INSERT INTO users (id, openid, balance, free_quota, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, id, "mock-openid", balance, 0, time.Now(), time.Now())
	return err
}

func (r *BillingRepo) DeductBalance(ctx context.Context, userID int64, amount float64) error {
	query := `UPDATE users SET balance = balance - $1, updated_at = $2 WHERE id = $3 AND balance >= $1`
	res, err := r.db.ExecContext(ctx, query, amount, time.Now(), userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *BillingRepo) AddBalance(ctx context.Context, userID int64, amount float64) error {
	query := `UPDATE users SET balance = balance + $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), userID)
	return err
}

func (r *BillingRepo) DecrementFreeQuota(ctx context.Context, userID int64) error {
	query := `UPDATE users SET free_quota = free_quota - 1, updated_at = $1 WHERE id = $2 AND free_quota > 0`
	res, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *BillingRepo) CreateTransaction(ctx context.Context, t *billing.Transaction) error {
	query := `
		INSERT INTO transactions (user_id, generation_id, type, amount, balance_before, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query,
		t.UserID, t.GenerationID, t.Type, t.Amount, t.BalanceBefore, t.BalanceAfter, time.Now(),
	).Scan(&t.ID)
}

func (r *BillingRepo) GetTransactionByGeneration(ctx context.Context, generationID int64) (*billing.Transaction, error) {
	query := `
		SELECT id, user_id, generation_id, type, amount, balance_before, balance_after
		FROM transactions
		WHERE generation_id = $1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, generationID)
	var t billing.Transaction
	err := row.Scan(&t.ID, &t.UserID, &t.GenerationID, &t.Type, &t.Amount, &t.BalanceBefore, &t.BalanceAfter)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *BillingRepo) CreateOrder(ctx context.Context, o *billing.Order) error {
	query := `
		INSERT INTO orders (user_id, order_no, package_code, amount, status, wx_prepay_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query,
		o.UserID, o.OrderNo, o.PackageCode, o.Amount, o.Status, o.WxOrderNo, time.Now(), time.Now(),
	).Scan(&o.ID)
}

func (r *BillingRepo) GetOrderByWxOrderNo(ctx context.Context, wxOrderNo string) (*billing.Order, error) {
	query := `
		SELECT id, user_id, order_no, package_code, amount, status, wx_prepay_id
		FROM orders
		WHERE wx_prepay_id = $1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, wxOrderNo)
	var o billing.Order
	var wxPrepayID sql.NullString
	err := row.Scan(&o.ID, &o.UserID, &o.OrderNo, &o.PackageCode, &o.Amount, &o.Status, &wxPrepayID)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	if wxPrepayID.Valid {
		o.WxOrderNo = wxPrepayID.String
	}
	return &o, nil
}

func (r *BillingRepo) UpdateOrderStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

// AtomicCharge performs balance deduction (or free-quota decrement) and transaction creation atomically.
func (r *BillingRepo) AtomicCharge(ctx context.Context, userID, generationID int64, amount float64, useFreeQuota bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM transactions WHERE generation_id = $1)", generationID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return billing.ErrGenerationAlreadyCharged
	}

	var balanceBefore, balanceAfter float64
	if useFreeQuota {
		res, err := tx.ExecContext(ctx, "UPDATE users SET free_quota = free_quota - 1, updated_at = $1 WHERE id = $2 AND free_quota > 0", time.Now(), userID)
		if err != nil {
			return err
		}
		if ra, _ := res.RowsAffected(); ra == 0 {
			return errors.New("insufficient free quota")
		}
		balanceBefore = 0
		balanceAfter = 0
		amount = 0
	} else {
		err = tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = $1", userID).Scan(&balanceBefore)
		if err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, "UPDATE users SET balance = balance - $1, updated_at = $2 WHERE id = $3 AND balance >= $1", amount, time.Now(), userID)
		if err != nil {
			return err
		}
		if ra, _ := res.RowsAffected(); ra == 0 {
			return errors.New("insufficient balance")
		}
		balanceAfter = balanceBefore - amount
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO transactions (user_id, generation_id, type, amount, balance_before, balance_after, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		userID, generationID, "generation_charge", amount, balanceBefore, balanceAfter, time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// AtomicRecharge updates order status to paid and adds balance atomically.
func (r *BillingRepo) AtomicRecharge(ctx context.Context, userID, orderID int64, amount float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, "UPDATE orders SET status = 'paid', updated_at = $1 WHERE id = $2 AND status != 'paid'", time.Now(), orderID)
	if err != nil {
		return err
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		// Already paid or not found; treat as idempotent success
		return nil
	}

	_, err = tx.ExecContext(ctx, "UPDATE users SET balance = balance + $1, updated_at = $2 WHERE id = $3", amount, time.Now(), userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
