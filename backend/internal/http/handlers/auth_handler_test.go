package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"image-play/internal/domain/user"
)

type mockUserRepo struct {
	nextID        int64
	usersByID     map[int64]*user.User
	usersByOpenID map[string]*user.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		nextID:        1,
		usersByID:     make(map[int64]*user.User),
		usersByOpenID: make(map[string]*user.User),
	}
}

func (r *mockUserRepo) GetByID(_ context.Context, id int64) (*user.User, error) {
	account, ok := r.usersByID[id]
	if !ok {
		return nil, nil
	}

	cloned := *account
	return &cloned, nil
}

func (r *mockUserRepo) GetByOpenID(_ context.Context, openID string) (*user.User, error) {
	account, ok := r.usersByOpenID[openID]
	if !ok {
		return nil, nil
	}

	cloned := *account
	return &cloned, nil
}

func (r *mockUserRepo) Create(_ context.Context, account *user.User) error {
	account.ID = r.nextID
	r.nextID++

	cloned := *account
	r.usersByID[account.ID] = &cloned
	r.usersByOpenID[account.OpenID] = &cloned
	return nil
}

func TestLoginReturnsTokenAndUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := user.NewService(newMockUserRepo())
	r := gin.New()
	r.POST("/api/auth/login", LoginHandler("test-secret", svc))

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

func TestLoginReturnsPersistedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := user.NewService(newMockUserRepo())
	r := gin.New()
	r.POST("/api/auth/login", LoginHandler("test-secret", svc))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"code":"wx-code-1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp LoginResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.AccessToken)
	require.Equal(t, "mock-openid-wx-code-1", resp.User.Openid)
	require.Equal(t, int64(3), resp.User.FreeQuota)
}
