package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"image-play/internal/infrastructure/llm"
)

type mockTextClient struct {
	chatFunc       func(ctx context.Context, messages []llm.Message) (string, error)
	chatStreamFunc func(ctx context.Context, messages []llm.Message) (llm.StreamReader, error)
}

func (m *mockTextClient) Chat(ctx context.Context, messages []llm.Message) (string, error) {
	return m.chatFunc(ctx, messages)
}

func (m *mockTextClient) ChatStream(ctx context.Context, messages []llm.Message) (llm.StreamReader, error) {
	return m.chatStreamFunc(ctx, messages)
}

type mockStreamReader struct {
	chunks []llm.Chunk
	idx    int
}

func (r *mockStreamReader) Recv() (llm.Chunk, error) {
	if r.idx >= len(r.chunks) {
		return llm.Chunk{}, io.EOF
	}
	c := r.chunks[r.idx]
	r.idx++
	return c, nil
}

func (r *mockStreamReader) Close() {}

func TestFaceReadingHandler_Stream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := &mockTextClient{
		chatStreamFunc: func(ctx context.Context, messages []llm.Message) (llm.StreamReader, error) {
			require.Len(t, messages, 1)
			require.Len(t, messages[0].Parts, 2)
			assert.Equal(t, llm.PartTypeText, messages[0].Parts[0].Type)
			assert.Equal(t, llm.PartTypeImage, messages[0].Parts[1].Type)
			return &mockStreamReader{
				chunks: []llm.Chunk{
					{Content: "hello"},
					{Content: " world"},
				},
			}, nil
		},
	}

	router := gin.New()
	router.POST("/api/face-reading", FaceReadingHandler(client))

	reqBody, _ := json.Marshal(map[string]string{
		"image_base64": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD",
	})
	req := httptest.NewRequest("POST", "/api/face-reading", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `{"chunk":"hello"}`)
	assert.Contains(t, body, `{"chunk":" world"}`)
	assert.Contains(t, body, "data: [DONE]")
}

func TestFaceReadingHandler_MissingBase64(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := &mockTextClient{}
	router := gin.New()
	router.POST("/api/face-reading", FaceReadingHandler(client))

	reqBody, _ := json.Marshal(map[string]string{
		"image_base64": "",
	})
	req := httptest.NewRequest("POST", "/api/face-reading", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFaceReadingHandler_StreamError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := &mockTextClient{
		chatStreamFunc: func(ctx context.Context, messages []llm.Message) (llm.StreamReader, error) {
			return nil, assert.AnError
		},
	}

	router := gin.New()
	router.POST("/api/face-reading", FaceReadingHandler(client))

	reqBody, _ := json.Marshal(map[string]string{
		"image_base64": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD",
	})
	req := httptest.NewRequest("POST", "/api/face-reading", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}
