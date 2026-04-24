package tracking

import "context"

type Service struct {
	repo Repository
}

type Repository interface {
	CreateEvent(ctx context.Context, userID int64, event string, payload map[string]any) error
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) TrackEvent(ctx context.Context, userID int64, event string, payload map[string]any) error {
	return s.repo.CreateEvent(ctx, userID, event, payload)
}
