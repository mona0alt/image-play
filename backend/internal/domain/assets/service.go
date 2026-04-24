package assets

import (
	"context"
	"fmt"
	"time"
)

type UploadIntentResponse struct {
	AssetID   int64  `json:"asset_id"`
	ObjectKey string `json:"object_key"`
	UploadURL string `json:"upload_url"`
	ExpireAt  string `json:"expire_at"`
}

type Repository interface {
	Create(ctx context.Context, asset *Asset) error
}

type Asset struct {
	ID        int64
	UserID    int64
	ObjectKey string
	URL       string
	CreatedAt time.Time
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUploadIntent(ctx context.Context, userID int64) (*UploadIntentResponse, error) {
	// For MVP: generate a mock presigned URL
	objectKey := fmt.Sprintf("uploads/%d/%d.jpg", userID, time.Now().Unix())
	asset := &Asset{
		UserID:    userID,
		ObjectKey: objectKey,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, asset); err != nil {
		return nil, err
	}

	expireAt := time.Now().Add(15 * time.Minute)
	resp := &UploadIntentResponse{
		AssetID:   asset.ID,
		ObjectKey: objectKey,
		UploadURL: fmt.Sprintf("https://mock-cos.example.com/%s?presigned=mock", objectKey),
		ExpireAt:  expireAt.Format(time.RFC3339),
	}
	return resp, nil
}
