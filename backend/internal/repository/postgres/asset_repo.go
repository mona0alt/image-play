package postgres

import (
	"context"
	"database/sql"
	"time"
)

type Asset struct {
	ID        int64
	UserID    int64
	ObjectKey string
	URL       string
	CreatedAt time.Time
}

type AssetRepo struct {
	db *sql.DB
}

func NewAssetRepo(db *sql.DB) *AssetRepo {
	return &AssetRepo{db: db}
}

func (r *AssetRepo) Create(ctx context.Context, asset *Asset) error {
	query := `
		INSERT INTO assets (user_id, object_key, url, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query, asset.UserID, asset.ObjectKey, asset.URL, asset.CreatedAt).Scan(&asset.ID)
}

func (r *AssetRepo) GetByID(ctx context.Context, id int64) (*Asset, error) {
	query := `SELECT id, user_id, object_key, url, created_at FROM assets WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var a Asset
	err := row.Scan(&a.ID, &a.UserID, &a.ObjectKey, &a.URL, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
