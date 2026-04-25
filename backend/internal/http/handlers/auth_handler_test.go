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

func TestLoginReusesPersistedUserForSameCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := user.NewService(newMockUserRepo())
	r := gin.New()
	r.POST("/api/auth/login", LoginHandler("test-secret", svc))

	firstReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"code":"wx-code-1"}`))
	firstReq.Header.Set("Content-Type", "application/json")
	firstResp := httptest.NewRecorder()
	r.ServeHTTP(firstResp, firstReq)

	secondReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"code":"wx-code-1"}`))
	secondReq.Header.Set("Content-Type", "application/json")
	secondResp := httptest.NewRecorder()
	r.ServeHTTP(secondResp, secondReq)

	require.Equal(t, http.StatusOK, firstResp.Code)
	require.Equal(t, http.StatusOK, secondResp.Code)

	var firstLogin LoginResponse
	var secondLogin LoginResponse
	require.NoError(t, json.Unmarshal(firstResp.Body.Bytes(), &firstLogin))
	require.NoError(t, json.Unmarshal(secondResp.Body.Bytes(), &secondLogin))
	require.Equal(t, firstLogin.User.ID, secondLogin.User.ID)
	require.Equal(t, firstLogin.User.Openid, secondLogin.User.Openid)
	require.Equal(t, int64(1), secondLogin.User.ID)
}

func TestMeReturnsUserFromUserRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	repo.usersByID[7] = &user.User{
		ID:        7,
		OpenID:    "mock-openid-wx-code-7",
		Balance:   12,
		FreeQuota: 2,
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(7))
		c.Next()
	})
	r.GET("/me", MeHandler(repo))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp User
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, int64(7), resp.ID)
	require.Equal(t, "mock-openid-wx-code-7", resp.Openid)
	require.Equal(t, int64(12), resp.Balance)
	require.Equal(t, int64(2), resp.FreeQuota)
}

func TestMeReturnsNotFoundWhenUserMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(7))
		c.Next()
	})
	r.GET("/me", MeHandler(newMockUserRepo()))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.JSONEq(t, `{"error":"user not found"}`, w.Body.String())
}
