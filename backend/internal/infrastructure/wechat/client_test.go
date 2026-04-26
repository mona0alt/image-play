package wechat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCode2SessionSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/sns/jscode2session", r.URL.Path)
		require.Equal(t, "app-id", r.URL.Query().Get("appid"))
		require.Equal(t, "app-secret", r.URL.Query().Get("secret"))
		require.Equal(t, "test-code", r.URL.Query().Get("js_code"))
		require.Equal(t, "authorization_code", r.URL.Query().Get("grant_type"))

		resp := Code2SessionResponse{
			OpenID:     "openid-123",
			SessionKey: "session-key-456",
			UnionID:    "unionid-789",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("app-id", "app-secret")
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.Code2Session(ctx, "test-code")
	require.NoError(t, err)
	require.Equal(t, "openid-123", result.OpenID)
	require.Equal(t, "session-key-456", result.SessionKey)
	require.Equal(t, "unionid-789", result.UnionID)
}

func TestCode2SessionWechatError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Code2SessionResponse{
			ErrCode: 40029,
			ErrMsg:  "invalid code",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("app-id", "app-secret")
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.Code2Session(ctx, "bad-code")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "WeChat error 40029")
}

func TestCode2SessionHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("app-id", "app-secret")
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.Code2Session(ctx, "code")
	require.Error(t, err)
	require.Nil(t, result)
}

func TestCode2SessionTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("app-id", "app-secret")
	client.baseURL = server.URL
	client.httpClient = &http.Client{Timeout: 1 * time.Millisecond}

	ctx := context.Background()
	result, err := client.Code2Session(ctx, "code")
	require.Error(t, err)
	require.Nil(t, result)
}
