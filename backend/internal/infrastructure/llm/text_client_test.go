package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextClient_Chat(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "hello world"}},
			},
		})
	}))
	defer mock.Close()

	client, err := NewTextClient(TextConfig{
		APIKey:  "fake-key",
		BaseURL: mock.URL,
		Model:   "test-model",
		Timeout: 10 * time.Second,
	})
	require.NoError(t, err)

	resp, err := client.Chat(context.Background(), []Message{
		{Role: "user", Parts: []Part{{Type: PartTypeText, Content: "hi"}}},
	})
	require.NoError(t, err)
	assert.Equal(t, "hello world", resp)
}

func TestTextClient_ChatStream(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		chunks := []string{"hello", " ", "world"}
		for _, c := range chunks {
			data, _ := json.Marshal(map[string]any{
				"choices": []map[string]any{
					{"delta": map[string]any{"content": c}},
				},
			})
			w.Write([]byte("data: " + string(data) + "\n\n"))
			flusher.Flush()
		}
		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer mock.Close()

	client, err := NewTextClient(TextConfig{
		APIKey:  "fake-key",
		BaseURL: mock.URL,
		Model:   "test-model",
		Timeout: 10 * time.Second,
	})
	require.NoError(t, err)

	reader, err := client.ChatStream(context.Background(), []Message{
		{Role: "user", Parts: []Part{{Type: PartTypeText, Content: "hi"}}},
	})
	require.NoError(t, err)
	defer reader.Close()

	var full string
	for {
		chunk, err := reader.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		full += chunk.Content
	}
	assert.Equal(t, "hello world", full)
}
