package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"image-play/internal/domain/generation"
)

func buildToken(userID int64, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": string(rune(userID)), // intentionally wrong for string conversion? No, use itoa style
	})
	// Actually use proper string
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "1",
	})
	s, _ := token.SignedString([]byte(secret))
	return s
}

func TestCreateGenerationSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := generation.NewInMemoryRepo()
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
	repo := generation.NewInMemoryRepo()
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
	repo := generation.NewInMemoryRepo()
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
