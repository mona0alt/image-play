# WeChat Login Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace mock login with real WeChat Mini Program login (jscode2session), generate auto-nicknames, and remove all mock code.

**Architecture:** Add a small WeChat HTTP client that calls `jscode2session`; replace the mock service method with one that uses this client; adjust auth handler, config, and frontend session layer.

**Tech Stack:** Go 1.22 + Gin, UniApp/Vue3, PostgreSQL

---

## File Structure

| File | Responsibility |
|------|---------------|
| `internal/config/config.go` | Add `WechatAppID` and `WechatAppSecret` fields |
| `internal/infrastructure/wechat/client.go` | NEW: HTTP client for WeChat `jscode2session` API |
| `internal/infrastructure/wechat/client_test.go` | NEW: Unit tests for WeChat client with HTTP mock |
| `internal/domain/user/service.go` | Replace `GetOrCreateByMockCode` with `GetOrCreateByWxCode`; auto-generate nickname |
| `internal/domain/user/service_test.go` | UPDATE: Replace mock-code tests with WeChat-code tests |
| `internal/http/handlers/auth_handler.go` | UPDATE: `LoginHandler` calls `GetOrCreateByWxCode`; JWT expiry 30d; remove `openid` from response |
| `internal/http/handlers/auth_handler_test.go` | UPDATE: Tests for new login flow |
| `cmd/api/main.go` | Inject WeChat client from config into router |
| `internal/http/router.go` | UPDATE: Accept WeChat client, pass to login handler |
| `frontend/src/services/session.ts` | REWRITE: Call `uni.login()` to get real code, send to backend |
| `frontend/src/App.vue` | UPDATE: Call `ensureSession()` instead of `ensureMockSession()` |
| `frontend/src/services/api.ts` | UPDATE: Remove `openid` from login return type |

---

### Task 1: Add WeChat config fields

**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: Add WeChat fields to Config struct**

```go
package config

import "os"

type Config struct {
	AppEnv          string
	Port            string
	DatabaseURL     string
	JWTSecret       string
	WechatAppID     string
	WechatAppSecret string
}

func Load() *Config {
	return &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		WechatAppID:     getEnv("WECHAT_APPID", ""),
		WechatAppSecret: getEnv("WECHAT_APPSECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/config/config.go
git commit -m "config: add wechat appid and appsecret fields"
```

---

### Task 2: Create WeChat HTTP client

**Files:**
- Create: `internal/infrastructure/wechat/client.go`
- Create: `internal/infrastructure/wechat/client_test.go`

- [ ] **Step 1: Write the failing test**

```go
package wechat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCode2SessionSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/sns/jscode2session", r.URL.Path)
		require.Equal(t, "test-appid", r.URL.Query().Get("appid"))
		require.Equal(t, "test-secret", r.URL.Query().Get("secret"))
		require.Equal(t, "wx-code-123", r.URL.Query().Get("js_code"))
		require.Equal(t, "authorization_code", r.URL.Query().Get("grant_type"))

		resp := map[string]interface{}{
			"openid":      "real-openid-123",
			"session_key": "session-key-abc",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-appid", "test-secret")
	client.baseURL = server.URL

	res, err := client.Code2Session(context.Background(), "wx-code-123")
	require.NoError(t, err)
	require.Equal(t, "real-openid-123", res.OpenID)
	require.Equal(t, "session-key-abc", res.SessionKey)
}

func TestCode2SessionReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"errcode": 40029,
			"errmsg":  "invalid code",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-appid", "test-secret")
	client.baseURL = server.URL

	_, err := client.Code2Session(context.Background(), "bad-code")
	require.Error(t, err)
	require.Contains(t, err.Error(), "wechat error")
}

func TestCode2SessionHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test-appid", "test-secret")
	client.baseURL = server.URL

	_, err := client.Code2Session(context.Background(), "wx-code")
	require.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/lili/Project/image-play/backend
go test ./internal/infrastructure/wechat/... -v
```

Expected: FAIL with "package not found" or "no go files"

- [ ] **Step 3: Write minimal implementation**

```go
package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://api.weixin.qq.com"

type Client struct {
	appID     string
	appSecret string
	baseURL   string
	httpClient *http.Client
}

type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func NewClient(appID, appSecret string) *Client {
	return &Client{
		appID:      appID,
		appSecret:  appSecret,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) Code2Session(ctx context.Context, code string) (*Code2SessionResponse, error) {
	u, _ := url.Parse(c.baseURL + "/sns/jscode2session")
	q := u.Query()
	q.Set("appid", c.appID)
	q.Set("secret", c.appSecret)
	q.Set("js_code", code)
	q.Set("grant_type", "authorization_code")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request wechat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wechat returned status %d", resp.StatusCode)
	}

	var result Code2SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error %d: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd /Users/lili/Project/image-play/backend
go test ./internal/infrastructure/wechat/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/wechat/
git commit -m "feat: add wechat jscode2session client"
```

---

### Task 3: Replace mock user service with WeChat login

**Files:**
- Modify: `internal/domain/user/service.go`
- Modify: `internal/domain/user/service_test.go`

- [ ] **Step 1: Write the failing test**

Replace contents of `internal/domain/user/service_test.go`:

```go
package user

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"image-play/internal/infrastructure/wechat"
)

type mockWxClient struct {
	openID     string
	sessionKey string
	err        error
}

func (m *mockWxClient) Code2Session(_ context.Context, _ string) (*wechat.Code2SessionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &wechat.Code2SessionResponse{OpenID: m.openID, SessionKey: m.sessionKey}, nil
}

type mockUserRepo struct {
	nextID        int64
	usersByID     map[int64]*User
	usersByOpenID map[string]*User
	createErr     error
	onCreate      func(*User)
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		nextID:        1,
		usersByID:     make(map[int64]*User),
		usersByOpenID: make(map[string]*User),
	}
}

func (r *mockUserRepo) GetByID(_ context.Context, id int64) (*User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return nil, nil
	}
	cloned := *user
	return &cloned, nil
}

func (r *mockUserRepo) GetByOpenID(_ context.Context, openID string) (*User, error) {
	user, ok := r.usersByOpenID[openID]
	if !ok {
		return nil, nil
	}
	cloned := *user
	return &cloned, nil
}

func (r *mockUserRepo) Create(_ context.Context, user *User) error {
	if r.onCreate != nil {
		r.onCreate(user)
	}
	if r.createErr != nil {
		return r.createErr
	}
	user.ID = r.nextID
	r.nextID++
	cloned := *user
	r.usersByID[user.ID] = &cloned
	r.usersByOpenID[user.OpenID] = &cloned
	return nil
}

func TestGetOrCreateByWxCodeCreatesUserOnFirstLogin(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-1", sessionKey: "sk-1"}

	user, created, err := svc.GetOrCreateByWxCode(context.Background(), "wx-code-1", wx)

	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, "wx-openid-1", user.OpenID)
	require.Equal(t, 3, user.FreeQuota)
	require.NotEmpty(t, user.Nickname)
	require.Contains(t, user.Nickname, "创作者")
}

func TestGetOrCreateByWxCodeReusesExistingUser(t *testing.T) {
	repo := newMockUserRepo()
	repo.usersByOpenID["wx-openid-1"] = &User{
		ID:        42,
		OpenID:    "wx-openid-1",
		Balance:   0,
		FreeQuota: 2,
		Nickname:  "创作者123",
	}
	svc := NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-1", sessionKey: "sk-1"}

	user, created, err := svc.GetOrCreateByWxCode(context.Background(), "wx-code-1", wx)

	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, int64(42), user.ID)
	require.Equal(t, 2, user.FreeQuota)
	require.Equal(t, "创作者123", user.Nickname)
}

func TestGetOrCreateByWxCodeReusesUserWhenCreateLosesRace(t *testing.T) {
	repo := newMockUserRepo()
	repo.createErr = errors.New("duplicate key value violates unique constraint")
	repo.onCreate = func(user *User) {
		repo.usersByID[99] = &User{
			ID:        99,
			OpenID:    user.OpenID,
			Balance:   0,
			FreeQuota: 3,
			Nickname:  user.Nickname,
		}
		repo.usersByOpenID[user.OpenID] = repo.usersByID[99]
	}
	svc := NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-1", sessionKey: "sk-1"}

	account, created, err := svc.GetOrCreateByWxCode(context.Background(), "wx-code-1", wx)

	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, int64(99), account.ID)
	require.Equal(t, "wx-openid-1", account.OpenID)
}

func TestGetOrCreateByWxCodeReturnsWeChatError(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)
	wx := &mockWxClient{err: errors.New("wechat api error")}

	_, _, err := svc.GetOrCreateByWxCode(context.Background(), "bad-code", wx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "wechat")
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/lili/Project/image-play/backend
go test ./internal/domain/user/... -v
```

Expected: FAIL with "GetOrCreateByWxCode not defined"

- [ ] **Step 3: Write minimal implementation**

Replace contents of `internal/domain/user/service.go`:

```go
package user

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"image-play/internal/infrastructure/wechat"
)

type User struct {
	ID        int64
	OpenID    string
	Balance   float64
	FreeQuota int
	Nickname  string
	AvatarURL string
}

type Repository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByOpenID(ctx context.Context, openID string) (*User, error)
	Create(ctx context.Context, user *User) error
}

type WechatClient interface {
	Code2Session(ctx context.Context, code string) (*wechat.Code2SessionResponse, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetOrCreateByWxCode(ctx context.Context, code string, wxClient WechatClient) (*User, bool, error) {
	wxRes, err := wxClient.Code2Session(ctx, code)
	if err != nil {
		return nil, false, fmt.Errorf("wechat login failed: %w", err)
	}

	existing, err := s.repo.GetByOpenID(ctx, wxRes.OpenID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	account := &User{
		OpenID:    wxRes.OpenID,
		Balance:   0,
		FreeQuota: 3,
		Nickname:  "创作者" + strconv.Itoa(rand.Intn(900)+100),
	}
	if err := s.repo.Create(ctx, account); err != nil {
		existing, getErr := s.repo.GetByOpenID(ctx, wxRes.OpenID)
		if getErr != nil {
			return nil, false, getErr
		}
		if existing != nil {
			return existing, false, nil
		}
		return nil, false, err
	}

	return account, true, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd /Users/lili/Project/image-play/backend
go test ./internal/domain/user/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/domain/user/
git commit -m "feat: replace mock login with wechat login in user service"
```

---

### Task 4: Update auth handler and tests

**Files:**
- Modify: `internal/http/handlers/auth_handler.go`
- Modify: `internal/http/handlers/auth_handler_test.go`

- [ ] **Step 1: Write the failing test**

Replace contents of `internal/http/handlers/auth_handler_test.go`:

```go
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
	"image-play/internal/infrastructure/wechat"
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

type mockWxClient struct {
	openID string
	err    error
}

func (m *mockWxClient) Code2Session(_ context.Context, _ string) (*wechat.Code2SessionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &wechat.Code2SessionResponse{OpenID: m.openID}, nil
}

func TestLoginReturnsTokenAndUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	svc := user.NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-new"}
	r := gin.New()
	r.POST("/api/auth/login", LoginHandler("test-secret", svc, wx))

	reqBody := `{"code":"wx-code-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ID        int64  `json:"id"`
			Balance   int64  `json:"balance"`
			FreeQuota int64  `json:"free_quota"`
			Nickname  string `json:"nickname"`
		} `json:"user"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.Equal(t, int64(1), resp.User.ID)
	require.Equal(t, int64(0), resp.User.Balance)
	require.Equal(t, int64(3), resp.User.FreeQuota)
	require.NotEmpty(t, resp.User.Nickname)
}

func TestLoginReusesPersistedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	repo.usersByOpenID["wx-openid-existing"] = &user.User{
		ID:        7,
		OpenID:    "wx-openid-existing",
		Balance:   12,
		FreeQuota: 2,
		Nickname:  "创作者888",
	}
	svc := user.NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-existing"}
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
	require.Equal(t, int64(7), resp.User.ID)
	require.Equal(t, int64(12), resp.User.Balance)
	require.Equal(t, int64(2), resp.User.FreeQuota)
	require.Equal(t, "创作者888", resp.User.Nickname)
}

func TestLoginReturnsWeChatError(t *testing.T) {
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
	require.Contains(t, w.Body.String(), "WECHAT_LOGIN_FAILED")
}

func TestMeReturnsUserFromUserRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	repo.usersByID[7] = &user.User{
		ID:        7,
		OpenID:    "wx-openid-7",
		Balance:   12,
		FreeQuota: 2,
		Nickname:  "创作者777",
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
	require.Equal(t, int64(12), resp.Balance)
	require.Equal(t, int64(2), resp.FreeQuota)
	require.Equal(t, "创作者777", resp.Nickname)
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/lili/Project/image-play/backend
go test ./internal/http/handlers/... -v -run TestLogin
```

Expected: FAIL with "LoginHandler signature mismatch"

- [ ] **Step 3: Write minimal implementation**

Replace contents of `internal/http/handlers/auth_handler.go`:

```go
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"image-play/internal/domain/user"
)

type LoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type User struct {
	ID        int64  `json:"id"`
	Balance   int64  `json:"balance"`
	FreeQuota int64  `json:"free_quota"`
	Nickname  string `json:"nickname"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

func LoginHandler(jwtSecret string, userSvc *user.Service, wxClient user.WechatClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		account, _, err := userSvc.GetOrCreateByWxCode(c.Request.Context(), req.Code, wxClient)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "WECHAT_LOGIN_FAILED", "error": "登录失败，请重试"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": strconv.FormatInt(account.ID, 10),
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
		})

		accessToken, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}

		c.JSON(http.StatusOK, LoginResponse{
			AccessToken: accessToken,
			User: User{
				ID:        account.ID,
				Balance:   int64(account.Balance),
				FreeQuota: int64(account.FreeQuota),
				Nickname:  account.Nickname,
			},
		})
	}
}

func MeHandler(userRepo user.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("user_id")
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		account, err := userRepo.GetByID(c.Request.Context(), uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
			return
		}
		if account == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, User{
			ID:        account.ID,
			Balance:   int64(account.Balance),
			FreeQuota: int64(account.FreeQuota),
			Nickname:  account.Nickname,
		})
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd /Users/lili/Project/image-play/backend
go test ./internal/http/handlers/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/http/handlers/
git commit -m "feat: update auth handler for wechat login, remove openid from response"
```

---

### Task 5: Update router to accept WeChat client

**Files:**
- Modify: `internal/http/router.go`

- [ ] **Step 1: Update router signature and login route**

Replace contents of `internal/http/router.go`:

```go
package http

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/assets"
	"image-play/internal/domain/billing"
	"image-play/internal/domain/generation"
	"image-play/internal/domain/tracking"
	"image-play/internal/domain/user"
	"image-play/internal/http/handlers"
	"image-play/internal/http/middleware"
	"image-play/internal/infrastructure/wechat"
	"image-play/internal/repository/postgres"
)

func NewRouter(db *sql.DB, jwtSecret string, wxClient *wechat.Client) *gin.Engine {
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	billingRepo := postgres.NewBillingRepo(db)
	billingSvc := billing.NewService(billingRepo)
	userRepo := postgres.NewUserRepo(db)
	userSvc := user.NewService(userRepo)
	templateRepo := postgres.NewSceneTemplateRepo(db)

	r.POST("/api/auth/login", handlers.LoginHandler(jwtSecret, userSvc, wxClient))
	r.GET("/api/configs/client", handlers.ClientConfigHandler(templateRepo))
	r.GET("/api/scenes/:scene_key/templates", handlers.ListSceneTemplatesHandler(templateRepo))

	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware(jwtSecret))
	authorized.GET("/me", handlers.MeHandler(userRepo))

	assetRepo := postgres.NewAssetRepo(db)
	assetSvc := assets.NewService(assetRepo)
	authorized.POST("/assets/upload-intent", handlers.UploadIntentHandler(assetSvc))

	genRepo := postgres.NewGenerationRepo(db)
	genSvc := generation.NewService(genRepo, templateRepo)
	authorized.POST("/generations", handlers.CreateGenerationHandler(genSvc))
	authorized.GET("/packages", handlers.PackagesHandler(billingSvc))
	authorized.POST("/orders", handlers.CreateOrderHandler(billingSvc))
	authorized.GET("/history", handlers.HistoryHandlerV2(genRepo))

	trackingRepo := postgres.NewTrackingRepo(db)
	trackingSvc := tracking.NewService(trackingRepo)
	authorized.POST("/tracking/events", handlers.TrackingEventsHandler(trackingSvc))

	authorized.GET("/explore/feed", handlers.ExploreFeedHandler(db))
	authorized.POST("/explore/like", handlers.ExploreLikeHandler(db))

	r.POST("/api/payments/callback", handlers.PaymentCallbackHandler(billingSvc))

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(jwtSecret))
	admin.GET("/metrics", handlers.DashboardMetricsHandler(db))
	admin.PUT("/templates/:id/toggle", handlers.ToggleTemplateHandler(db))

	return r
}
```

- [ ] **Step 2: Update main.go to create WeChat client and pass to router**

Replace contents of `cmd/api/main.go`:

```go
package main

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"image-play/internal/config"
	http "image-play/internal/http"
	"image-play/internal/infrastructure/wechat"
	"image-play/internal/migration"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if cfg.WechatAppID == "" || cfg.WechatAppSecret == "" {
		log.Println("[warning] WECHAT_APPID or WECHAT_APPSECRET not set, login will fail")
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migration.Run(context.Background(), db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	wxClient := wechat.NewClient(cfg.WechatAppID, cfg.WechatAppSecret)
	r := http.NewRouter(db, cfg.JWTSecret, wxClient)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
```

- [ ] **Step 3: Run backend tests**

```bash
cd /Users/lili/Project/image-play/backend
go test ./... -v
```

Expected: PASS (all tests)

- [ ] **Step 4: Commit**

```bash
git add internal/http/router.go cmd/api/main.go
git commit -m "feat: wire wechat client into router and main"
```

---

### Task 6: Rewrite frontend session service

**Files:**
- Modify: `frontend/src/services/session.ts`

- [ ] **Step 1: Rewrite session.ts**

```ts
import { login } from './api'

export async function ensureSession(): Promise<string> {
  const existing = uni.getStorageSync('access_token') as string | undefined
  if (existing) {
    return existing
  }

  const [err, res] = await uni.login({ provider: 'weixin' })
  if (err || !res || !res.code) {
    throw new Error('微信登录失败')
  }

  const loginRes = await login(res.code)
  uni.setStorageSync('access_token', loginRes.access_token)
  return loginRes.access_token
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/services/session.ts
git commit -m "feat: rewrite session service for real wechat login"
```

---

### Task 7: Update App.vue to use new session function

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Update import and call**

```vue
<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { ensureSession } from './services/session'

onLaunch(() => {
  void ensureSession()
})
</script>

<style>
page {
  --gallery-bg: #fdf8f8;
  --gallery-surface: #ffffff;
  --gallery-surface-soft: #f4efed;
  --gallery-border: rgba(28, 27, 27, 0.08);
  --gallery-text: #1c1b1b;
  --gallery-muted: #6d6865;
  --gallery-accent: #111111;
  background: var(--gallery-bg);
  color: var(--gallery-text);
}

view,
text,
button,
image,
scroll-view,
input,
textarea {
  box-sizing: border-box;
}

button {
  border: none;
}

button::after {
  border: none;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/App.vue
git commit -m "feat: update App.vue to call ensureSession"
```

---

### Task 8: Update frontend API types

**Files:**
- Modify: `frontend/src/services/api.ts`

- [ ] **Step 1: Remove openid from login return type, add nickname**

In `frontend/src/services/api.ts`, update the `login` function signature:

```ts
export function login(code: string) {
  return request<{ access_token: string; user: { id: number; balance: number; free_quota: number; nickname: string } }>({
    url: '/api/auth/login',
    method: 'POST',
    data: { code },
    headers: { 'Content-Type': 'application/json' },
  })
}
```

And update the `getMe` return type:

```ts
export function getMe() {
  return request<{ id: number; balance: number; free_quota: number; nickname: string }>({
    url: '/api/me',
    method: 'GET',
  })
}
```

- [ ] **Step 2: Update UserProfile interface in store**

In `frontend/src/store/user.ts`, update the interface:

```ts
export interface UserProfile {
  id: number
  balance: number
  free_quota: number
  nickname: string
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/services/api.ts frontend/src/store/user.ts
git commit -m "feat: remove openid from api types, add nickname"
```

---

### Task 9: Full test run and final verification

- [ ] **Step 1: Run backend tests**

```bash
cd /Users/lili/Project/image-play/backend
go test ./... -v
```

Expected: ALL PASS

- [ ] **Step 2: Check frontend TypeScript**

```bash
cd /Users/lili/Project/image-play/frontend
npm run type-check
```

Expected: No type errors

- [ ] **Step 3: Final commit**

```bash
git commit --allow-empty -m "feat: complete wechat login integration"
```

---

## Self-Review

**1. Spec coverage:**
- [x] Config with appid/appsecret → Task 1
- [x] WeChat client calling jscode2session → Task 2
- [x] GetOrCreateByWxCode replacing mock → Task 3
- [x] Auto-generated nickname (创作者 + 3 digits) → Task 3
- [x] JWT 30 days → Task 4
- [x] Remove openid from response → Task 4
- [x] Frontend uni.login → Task 6
- [x] Error handling (WECHAT_LOGIN_FAILED) → Task 4

**2. Placeholder scan:** No TBD, TODO, or vague steps found.

**3. Type consistency:**
- `User` struct in handler uses `Nickname string` consistently
- `LoginResponse.User` matches `MeHandler` response type
- Frontend `UserProfile` includes `nickname`
- `wechat.Client` interface matches `user.WechatClient`

**4. Gap found and fixed:**
- Added `internal/http/router.go` update to Task 5 (was missing in initial file list)
- Added frontend `store/user.ts` update to Task 8
