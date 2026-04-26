package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.weixin.qq.com"

type Code2SessionResult struct {
	OpenID     string
	SessionKey string
	UnionID    string
	ErrCode    int
	ErrMsg     string
}

type Code2SessionResponse = Code2SessionResult

type Client struct {
	appID      string
	appSecret  string
	baseURL    string
	httpClient *http.Client
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
	url := fmt.Sprintf("%s/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		c.baseURL, c.appID, c.appSecret, code)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wechat API returned status %d", resp.StatusCode)
	}

	var result Code2SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("WeChat error %d: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
