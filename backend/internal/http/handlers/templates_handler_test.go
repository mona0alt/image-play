package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"image-play/internal/domain/scenes"
)

type mockSceneTemplateRepo struct {
	templates []scenes.Template
	err       error
}

func (r *mockSceneTemplateRepo) ListActiveByScene(_ context.Context, sceneKey string) ([]scenes.Template, error) {
	if r.err != nil {
		return nil, r.err
	}
	var items []scenes.Template
	for _, template := range r.templates {
		if template.SceneKey == sceneKey {
			items = append(items, template)
		}
	}
	return items, nil
}

func TestListSceneTemplatesReturnsOnlyActiveTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockSceneTemplateRepo{
		templates: []scenes.Template{
			{
				SceneKey:       scenes.ScenePortrait,
				TemplateKey:    "office-pro",
				Name:           "通勤职业",
				FormSchema:     scenes.FormSchema{{Name: "subject_name", Label: "拍摄对象", Type: "text", Required: true}},
				SampleImageURL: "https://example.com/portrait-office-pro.png",
				Active:         true,
			},
			{
				SceneKey:    scenes.ScenePortrait,
				TemplateKey: "disabled",
				Name:        "停用模板",
				Active:      false,
			},
		},
	}

	r := gin.New()
	r.GET("/api/scenes/:scene_key/templates", ListSceneTemplatesHandler(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/scenes/portrait/templates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{
		"items":[
			{
				"key":"office-pro",
				"name":"通勤职业",
				"scene_key":"portrait",
				"form_schema":[{"name":"subject_name","label":"拍摄对象","type":"text","required":true}],
				"sample_image_url":"https://example.com/portrait-office-pro.png"
			}
		]
	}`, w.Body.String())
}

func TestClientConfigReturnsSceneHallOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ClientConfigHandler(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{
		"brand_slogan":"Play with your images",
		"pricing":{"single":"1.00","pack10":"8.00"},
		"scene_order":["portrait","festival","invitation","tshirt","poster"]
	}`, w.Body.String())
}
