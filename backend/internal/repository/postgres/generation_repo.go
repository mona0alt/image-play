package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"image-play/internal/domain/generation"
)

type GenerationRepo struct {
	db *sql.DB
}

func NewGenerationRepo(db *sql.DB) *GenerationRepo {
	return &GenerationRepo{db: db}
}

func (r *GenerationRepo) Create(ctx context.Context, g *generation.Generation) error {
	fieldsJSON, err := json.Marshal(g.Fields)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO generations (user_id, client_request_id, scene_key, template_key, fields, source_asset_id, status, result_url, prompt, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query,
		g.UserID, g.ClientRequestID, g.SceneKey, g.TemplateKey, fieldsJSON, g.SourceAssetID, g.Status, g.ResultURL, g.Prompt, g.CreatedAt, g.UpdatedAt,
	).Scan(&g.ID)
}

func (r *GenerationRepo) GetActiveByUser(ctx context.Context, userID int64) (*generation.Generation, error) {
	query := `
		SELECT id, user_id, client_request_id, scene_key, template_key, fields, source_asset_id, status, result_url, prompt, created_at, updated_at
		FROM generations
		WHERE user_id = $1 AND status IN ('queued', 'running', 'result_auditing')
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, userID)
	return scanGeneration(row)
}

func (r *GenerationRepo) Dequeue(ctx context.Context) (*generation.Generation, error) {
	// Simple polling dequeue using SELECT ... FOR UPDATE SKIP LOCKED would be ideal,
	// but for MVP we use a simple select + update pattern.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT id, user_id, client_request_id, scene_key, template_key, fields, source_asset_id, status, result_url, prompt, created_at, updated_at
		FROM generations
		WHERE status = 'queued'
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`
	g, err := scanGeneration(tx.QueryRowContext(ctx, query))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE generations SET status = 'running', updated_at = $1 WHERE id = $2`, time.Now(), g.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	g.Status = "running"
	return g, nil
}

func (r *GenerationRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE generations SET status = $1, updated_at = $2 WHERE id = $3`, status, time.Now(), id)
	return err
}

func (r *GenerationRepo) UpdateResult(ctx context.Context, id int64, status, resultURL string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE generations SET status = $1, result_url = $2, updated_at = $3 WHERE id = $4`, status, resultURL, time.Now(), id)
	return err
}

func scanGeneration(row *sql.Row) (*generation.Generation, error) {
	var g generation.Generation
	var fieldsRaw []byte
	var sourceAssetID sql.NullInt64
	err := row.Scan(
		&g.ID, &g.UserID, &g.ClientRequestID, &g.SceneKey, &g.TemplateKey,
		&fieldsRaw, &sourceAssetID, &g.Status, &g.ResultURL, &g.Prompt,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if sourceAssetID.Valid {
		g.SourceAssetID = &sourceAssetID.Int64
	}
	if len(fieldsRaw) > 0 {
		if err := json.Unmarshal(fieldsRaw, &g.Fields); err != nil {
			return nil, err
		}
	}
	return &g, nil
}
