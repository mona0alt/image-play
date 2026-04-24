package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"image-play/internal/domain/generation"
)

func buildToken(userID int64, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.FormatInt(userID, 10),
	})
	s, _ := token.SignedString([]byte(secret))
	return s
}

type mockGenerationRepo struct {
	mu          sync.Mutex
	generations map[int64]*generation.Generation
	nextID      int64
}

func newMockGenerationRepo() *mockGenerationRepo {
	return &mockGenerationRepo{
		generations: make(map[int64]*generation.Generation),
		nextID:      1,
	}
}

func (r *mockGenerationRepo) Create(_ context.Context, g *generation.Generation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.generations {
		if existing.UserID == g.UserID && (existing.Status == "queued" || existing.Status == "running" || existing.Status == "result_auditing") {
			return errors.New("pq: duplicate key value violates unique constraint \"unique_active_generation_per_user\"")
		}
	}
	g.ID = r.nextID
	r.nextID++
	r.generations[g.ID] = g
	return nil
}

func (r *mockGenerationRepo) GetActiveByUser(_ context.Context, userID int64) (*generation.Generation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, g := range r.generations {
		if g.UserID == userID && (g.Status == "queued" || g.Status == "running" || g.Status == "result_auditing") {
			return g, nil
		}
	}
	return nil, nil
}

func (r *mockGenerationRepo) Dequeue(_ context.Context) (*generation.Generation, error) {
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

func (r *mockGenerationRepo) UpdateStatus(_ context.Context, id int64, status string) error {
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

func (r *mockGenerationRepo) UpdateResult(_ context.Context, id int64, status, resultURL string) error {
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

func TestCreateGenerationSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockGenerationRepo()
	svc := generation.NewService(repo)
	r := gin.New()
	r.POST("/api/generations", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		CreateGenerationHandler(svc)(c)
	})

	body, _ := json.Marshal(CreateGenerationRequest{
		ClientRequestID: "req-1",
		SceneKey:        "portrait",
		TemplateKey:     "office-pro",
		Fields:          map[string]string{},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateGenerationRejectsActiveJob(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockGenerationRepo()
	svc := generation.NewService(repo)
	r := gin.New()
	r.POST("/api/generations", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		CreateGenerationHandler(svc)(c)
	})

	// First request succeeds
	body1, _ := json.Marshal(CreateGenerationRequest{
		ClientRequestID: "req-1",
		SceneKey:        "portrait",
		TemplateKey:     "office-pro",
		Fields:          map[string]string{},
	})
	req1 := httptest.NewRequest(http.MethodPost, "/api/generations", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusCreated, w1.Code)

	// Second request conflicts
	body2, _ := json.Marshal(CreateGenerationRequest{
		ClientRequestID: "req-2",
		SceneKey:        "invitation",
		TemplateKey:     "wedding-classic",
		Fields:          map[string]string{},
	})
	req2 := httptest.NewRequest(http.MethodPost, "/api/generations", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusConflict, w2.Code)
}

func TestCreateGenerationBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockGenerationRepo()
	svc := generation.NewService(repo)
	r := gin.New()
	r.POST("/api/generations", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		CreateGenerationHandler(svc)(c)
	})

	body, _ := json.Marshal(map[string]string{
		"scene_key": "portrait",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
