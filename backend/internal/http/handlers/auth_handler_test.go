package handlers

import (
	"bytes"
	"encoding/json"
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

	var resp struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ID     int64  `json:"id"`
			OpenID string `json:"openid"`
		} `json:"user"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.Equal(t, int64(1), resp.User.ID)
	require.Equal(t, "mock-openid-mock-wechat-code", resp.User.OpenID)
}
