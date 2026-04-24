package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestLoginReturnsTokenAndUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// register handler
	r.POST("/api/auth/login", LoginHandler("test-secret"))

	reqBody := `{"code":"mock-wechat-code"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "access_token")
	require.Contains(t, w.Body.String(), "user")
}
