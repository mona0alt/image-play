package generation

import (
	"context"
	"errors"
	"strings"
	"time"

	"image-play/internal/domain/scenes"
)

var ErrActiveGenerationExists = errors.New("active generation already exists")
var ErrTemplateNotAvailable = errors.New("template not available")

var activeStatuses = map[string]bool{
	"queued":          true,
	"running":         true,
	"result_auditing": true,
}

type CreateGenerationInput struct {
	UserID          int64
	ClientRequestID string
	SceneKey        string
	TemplateKey     string
	Fields          map[string]string
	SourceAssetID   *int64
}

type Generation struct {
	ID              int64
	UserID          int64
	ClientRequestID string
	SceneKey        string
	TemplateKey     string
	Fields          map[string]string
	SourceAssetID   *int64
	Status          string
	ResultURL       string
	Prompt          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Repository interface {
	Create(ctx context.Context, g *Generation) error
	GetActiveByUser(ctx context.Context, userID int64) (*Generation, error)
	Dequeue(ctx context.Context) (*Generation, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdateResult(ctx context.Context, id int64, status, resultURL string) error
	ListByUser(ctx context.Context, userID int64) ([]*Generation, error)
}

type TemplateLookup interface {
	GetActiveTemplate(ctx context.Context, sceneKey, templateKey string) (*scenes.Template, error)
}

type Service struct {
	repo      Repository
	templates TemplateLookup
}

func NewService(repo Repository, templates TemplateLookup) *Service {
	return &Service{repo: repo, templates: templates}
}

func (s *Service) CreateGeneration(ctx context.Context, input CreateGenerationInput) (int64, error) {
	if s.templates == nil {
		return 0, errors.New("template lookup not configured")
	}

	template, err := s.templates.GetActiveTemplate(ctx, input.SceneKey, input.TemplateKey)
	if err != nil {
		return 0, err
	}
	if template == nil {
		return 0, ErrTemplateNotAvailable
	}

	now := time.Now()
	g := &Generation{
		UserID:          input.UserID,
		ClientRequestID: input.ClientRequestID,
		SceneKey:        input.SceneKey,
		TemplateKey:     input.TemplateKey,
		Fields:          input.Fields,
		SourceAssetID:   input.SourceAssetID,
		Status:          "queued",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Create(ctx, g); err != nil {
		if isUniqueViolation(err) {
			return 0, ErrActiveGenerationExists
		}
		return 0, err
	}
	return g.ID, nil
}

func isUniqueViolation(err error) bool {
	// Check for PostgreSQL unique violation (error code 23505)
	// This is a simplified check; in production you might use github.com/lib/pq
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "unique constraint")
}
