package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClientConfig struct {
	BrandSlogan string            `json:"brand_slogan"`
	Pricing     map[string]string `json:"pricing"`
	SceneOrder  []string          `json:"scene_order"`
}

func ClientConfigHandler(c *gin.Context) {
	config := ClientConfig{
		BrandSlogan: "Play with your images",
		Pricing: map[string]string{
			"single": "1.00",
			"pack10": "8.00",
		},
		SceneOrder: []string{"portrait", "landscape", "fun"},
	}
	c.JSON(http.StatusOK, config)
}
