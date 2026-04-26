package postgres

import (
	"context"
	"database/sql"
	"time"

	"image-play/internal/domain/user"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {
	const query = `SELECT id, openid, balance, free_quota, nickname, avatar_url FROM users WHERE id = $1`

	account := &user.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.OpenID,
		&account.Balance,
		&account.FreeQuota,
		&account.Nickname,
		&account.AvatarURL,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *UserRepo) GetByOpenID(ctx context.Context, openID string) (*user.User, error) {
	const query = `SELECT id, openid, balance, free_quota, nickname, avatar_url FROM users WHERE openid = $1`

	account := &user.User{}
	err := r.db.QueryRowContext(ctx, query, openID).Scan(
		&account.ID,
		&account.OpenID,
		&account.Balance,
		&account.FreeQuota,
		&account.Nickname,
		&account.AvatarURL,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *UserRepo) Create(ctx context.Context, account *user.User) error {
	const query = `
		INSERT INTO users (openid, balance, free_quota, nickname, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	now := time.Now()
	return r.db.QueryRowContext(ctx, query,
		account.OpenID,
		account.Balance,
		account.FreeQuota,
		account.Nickname,
		now,
		now,
	).Scan(&account.ID)
}

func (r *UserRepo) UpdateNickname(ctx context.Context, id int64, nickname string) error {
	const query = `UPDATE users SET nickname = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, nickname, time.Now(), id)
	return err
}
