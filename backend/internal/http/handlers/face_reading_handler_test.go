package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFaceReadingHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDMX := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "这是一段面相分析结果。",
					},
				},
			},
		})
	}))
	defer mockDMX.Close()

	router := gin.New()
	router.POST("/api/face-reading", FaceReadingHandler("fake-key", mockDMX.URL, "kimi-k2.6"))

	reqBody, _ := json.Marshal(map[string]string{
		"image_base64": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD",
	})
	req := httptest.NewRequest("POST", "/api/face-reading", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "这是一段面相分析结果。", resp["result"])
}

func TestFaceReadingHandler_MissingBase64(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/api/face-reading", FaceReadingHandler("fake-key", "http://localhost", "model"))

	reqBody, _ := json.Marshal(map[string]string{
		"image_base64": "",
	})
	req := httptest.NewRequest("POST", "/api/face-reading", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFaceReadingHandler_DMXFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDMX := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockDMX.Close()

	router := gin.New()
	router.POST("/api/face-reading", FaceReadingHandler("fake-key", mockDMX.URL, "model"))

	reqBody, _ := json.Marshal(map[string]string{
		"image_base64": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD",
	})
	req := httptest.NewRequest("POST", "/api/face-reading", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}
