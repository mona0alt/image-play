package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageClient_Generate(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer fake-key", r.Header.Get("Authorization"))

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "test-prompt", body["prompt"])
		assert.Equal(t, "test-model", body["model"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"url": "https://example.com/image.png"},
			},
		})
	}))
	defer mock.Close()

	client, err := NewImageClient(ImageConfig{
		APIKey:  "fake-key",
		BaseURL: mock.URL,
		Model:   "test-model",
		Timeout: 10 * time.Second,
	})
	require.NoError(t, err)

	url, err := client.Generate(context.Background(), "test-prompt")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/image.png", url)
}

func TestImageClient_Generate_APIError(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer mock.Close()

	client, err := NewImageClient(ImageConfig{
		APIKey:  "fake-key",
		BaseURL: mock.URL,
		Model:   "test-model",
		Timeout: 10 * time.Second,
	})
	require.NoError(t, err)

	_, err = client.Generate(context.Background(), "test-prompt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image API error")
}
