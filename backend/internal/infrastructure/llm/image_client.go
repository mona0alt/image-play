package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type imageClient struct {
	cfg    ImageConfig
	client *http.Client
}

func NewImageClient(cfg ImageConfig) (ImageClient, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &imageClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *imageClient) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody, err := json.Marshal(map[string]any{
		"model":  c.cfg.Model,
		"prompt": prompt,
	})
	if err != nil {
		return "", fmt.Errorf("llm: marshal image request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("llm: build image request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm: image request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm: image API error status=%d body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("llm: decode image response: %w", err)
	}
	if len(result.Data) == 0 || result.Data[0].URL == "" {
		return "", fmt.Errorf("llm: image response contains no url")
	}
	return result.Data[0].URL, nil
}
