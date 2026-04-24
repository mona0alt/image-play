package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
)

type TrackingRepo struct {
	db *sql.DB
}

func NewTrackingRepo(db *sql.DB) *TrackingRepo {
	return &TrackingRepo{db: db}
}

func (r *TrackingRepo) CreateEvent(ctx context.Context, userID int64, event string, payload map[string]any) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, "INSERT INTO tracking_events (user_id, event, payload) VALUES ($1, $2, $3)", userID, event, payloadJSON)
	return err
}
