package generation

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrActiveGenerationExists = errors.New("active generation already exists")

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
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateGeneration(ctx context.Context, input CreateGenerationInput) error {
	active, err := s.repo.GetActiveByUser(ctx, input.UserID)
	if err != nil {
		return err
	}
	if active != nil {
		return ErrActiveGenerationExists
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
	return s.repo.Create(ctx, g)
}

type InMemoryRepo struct {
	mu          sync.Mutex
	generations map[int64]*Generation
	nextID      int64
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		generations: make(map[int64]*Generation),
		nextID:      1,
	}
}

func (r *InMemoryRepo) Create(_ context.Context, g *Generation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g.ID = r.nextID
	r.nextID++
	r.generations[g.ID] = g
	return nil
}

func (r *InMemoryRepo) GetActiveByUser(_ context.Context, userID int64) (*Generation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, g := range r.generations {
		if g.UserID == userID && activeStatuses[g.Status] {
			return g, nil
		}
	}
	return nil, nil
}

func (r *InMemoryRepo) Dequeue(_ context.Context) (*Generation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, g := range r.generations {
		if g.Status == "queued" {
			g.Status = "running"
			g.UpdatedAt = time.Now()
			return g, nil
		}
	}
	return nil, nil
}

func (r *InMemoryRepo) UpdateStatus(_ context.Context, id int64, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.generations[id]
	if !ok {
		return errors.New("generation not found")
	}
	g.Status = status
	g.UpdatedAt = time.Now()
	return nil
}
