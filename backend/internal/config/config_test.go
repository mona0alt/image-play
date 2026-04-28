package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFromEnv_DMXConfig(t *testing.T) {
	os.Setenv("DMX_API_KEY", "test-key")
	os.Setenv("DMX_API_BASE_URL", "https://test.example.com/v1")
	os.Setenv("DMX_MODEL", "test-model")
	defer func() {
		os.Unsetenv("DMX_API_KEY")
		os.Unsetenv("DMX_API_BASE_URL")
		os.Unsetenv("DMX_MODEL")
	}()

	cfg := loadFromEnv()
	assert.Equal(t, "test-key", cfg.DMXAPIKey)
	assert.Equal(t, "https://test.example.com/v1", cfg.DMXAPIBaseURL)
	assert.Equal(t, "test-model", cfg.DMXModel)
}
