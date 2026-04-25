package generation

import (
	"context"
	"errors"
	"sync"
	"time"

	"image-play/internal/domain/scenes"
)

type inMemoryRepo struct {
	mu          sync.Mutex
	generations map[int64]*Generation
	nextID      int64
}

func newInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{
		generations: make(map[int64]*Generation),
		nextID:      1,
	}
}

type inMemoryTemplateRepo struct {
	mu        sync.Mutex
	templates map[string]*scenes.Template
	err       error
}

func newInMemoryTemplateRepo() *inMemoryTemplateRepo {
	return &inMemoryTemplateRepo{
		templates: make(map[string]*scenes.Template),
	}
}

func (r *inMemoryTemplateRepo) Set(template *scenes.Template) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if template == nil {
		return
	}
	r.templates[templateLookupKey(template.SceneKey, template.TemplateKey)] = template
}

func (r *inMemoryTemplateRepo) GetActiveTemplate(_ context.Context, sceneKey, templateKey string) (*scenes.Template, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return nil, r.err
	}
	return r.templates[templateLookupKey(sceneKey, templateKey)], nil
}

func templateLookupKey(sceneKey, templateKey string) string {
	return sceneKey + ":" + templateKey
}

func (r *inMemoryRepo) Create(_ context.Context, g *Generation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.generations {
		if existing.UserID == g.UserID && activeStatuses[existing.Status] {
			return errors.New("pq: duplicate key value violates unique constraint \"unique_active_generation_per_user\"")
		}
	}
	g.ID = r.nextID
	r.nextID++
	r.generations[g.ID] = g
	return nil
}

func (r *inMemoryRepo) GetActiveByUser(_ context.Context, userID int64) (*Generation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, g := range r.generations {
		if g.UserID == userID && activeStatuses[g.Status] {
			return g, nil
		}
	}
	return nil, nil
}

func (r *inMemoryRepo) Dequeue(_ context.Context) (*Generation, error) {
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

func (r *inMemoryRepo) UpdateStatus(_ context.Context, id int64, status string) error {
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

func (r *inMemoryRepo) UpdateResult(_ context.Context, id int64, status, resultURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.generations[id]
	if !ok {
		return errors.New("generation not found")
	}
	g.Status = status
	g.ResultURL = resultURL
	g.UpdatedAt = time.Now()
	return nil
}

func (r *inMemoryRepo) ListByUser(_ context.Context, userID int64) ([]*Generation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var results []*Generation
	for _, g := range r.generations {
		if g.UserID == userID {
			results = append(results, g)
		}
	}
	return results, nil
}
