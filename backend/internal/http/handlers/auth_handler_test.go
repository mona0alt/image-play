package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"image-play/internal/domain/user"
	"image-play/internal/infrastructure/wechat"
)

type mockWxClient struct {
	openID string
	err    error
}

func (m *mockWxClient) Code2Session(_ context.Context, _ string) (*wechat.Code2SessionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &wechat.Code2SessionResponse{
		OpenID:     m.openID,
		SessionKey: "session-key",
	}, nil
}

type mockUserRepo struct {
	nextID        int64
	usersByID     map[int64]*user.User
	usersByOpenID map[string]*user.User
	createErr     error
	onCreate      func(*user.User)
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
	if r.onCreate != nil {
		r.onCreate(account)
	}
	if r.createErr != nil {
		return r.createErr
	}

	account.ID = r.nextID
	r.nextID++

	cloned := *account
	r.usersByID[account.ID] = &cloned
	r.usersByOpenID[account.OpenID] = &cloned
	return nil
}

func (r *mockUserRepo) UpdateNickname(_ context.Context, id int64, nickname string) error {
	if account, ok := r.usersByID[id]; ok {
		account.Nickname = nickname
	}
	return nil
}

func TestLoginReturnsTokenAndUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	svc := user.NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-1"}
	r := gin.New()
	r.POST("/api/auth/login", LoginHandler("test-secret", svc, wx))

	reqBody := `{"code":"wx-code-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.Equal(t, int64(1), resp.User.ID)
	require.NotEmpty(t, resp.User.Nickname)
	require.Equal(t, int64(3), resp.User.FreeQuota)
}

func TestLoginReturnsPersistedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	repo.usersByOpenID["wx-openid-1"] = &user.User{
		ID:        7,
		OpenID:    "wx-openid-1",
		Balance:   12,
		FreeQuota: 2,
		Nickname:  "ExistingUser",
	}
	svc := user.NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-1"}
	r := gin.New()
	r.POST("/api/auth/login", LoginHandler("test-secret", svc, wx))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"code":"wx-code-1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp LoginResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.AccessToken)
	require.Equal(t, "ExistingUser", resp.User.Nickname)
	require.Equal(t, int64(7), resp.User.ID)
	require.Equal(t, int64(2), resp.User.FreeQuota)
}

func TestLoginReusesPersistedUserForSameCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	svc := user.NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-1"}
	r := gin.New()
	r.POST("/api/auth/login", LoginHandler("test-secret", svc, wx))

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
	require.Equal(t, firstLogin.User.Nickname, secondLogin.User.Nickname)
	require.Equal(t, int64(1), secondLogin.User.ID)
}

func TestLoginHandlesWechatError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	svc := user.NewService(repo)
	wx := &mockWxClient{err: errors.New("wechat api error")}
	r := gin.New()
	r.POST("/api/auth/login", LoginHandler("test-secret", svc, wx))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"code":"bad-code"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.JSONEq(t, `{"code":"WECHAT_LOGIN_FAILED","error":"登录失败，请重试"}`, w.Body.String())
}

func TestMeReturnsUserFromUserRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	repo.usersByID[7] = &user.User{
		ID:        7,
		OpenID:    "wx-openid-7",
		Balance:   12,
		FreeQuota: 2,
		Nickname:  "TestUser",
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
	require.Equal(t, "TestUser", resp.Nickname)
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

func TestUpdateMeUpdatesNickname(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	repo.usersByID[7] = &user.User{
		ID:        7,
		OpenID:    "wx-openid-7",
		Balance:   12,
		FreeQuota: 2,
		Nickname:  "OldName",
	}
	svc := user.NewService(repo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(7))
		c.Next()
	})
	r.PUT("/me", UpdateMeHandler(svc))

	reqBody := `{"nickname":"NewName"}`
	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp User
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, int64(7), resp.ID)
	require.Equal(t, "NewName", resp.Nickname)
	require.Equal(t, int64(12), resp.Balance)
	require.Equal(t, int64(2), resp.FreeQuota)

	// Verify persisted
	updated, _ := repo.GetByID(context.Background(), 7)
	require.Equal(t, "NewName", updated.Nickname)
}

func TestUpdateMeReturnsUnauthorizedWhenMissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := user.NewService(newMockUserRepo())
	r := gin.New()
	r.PUT("/me", UpdateMeHandler(svc))

	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(`{"nickname":"Name"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMeReturnsBadRequestForInvalidNickname(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	repo.usersByID[7] = &user.User{ID: 7}
	svc := user.NewService(repo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(7))
		c.Next()
	})
	r.PUT("/me", UpdateMeHandler(svc))

	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(`{"nickname":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
