# LLM 通用模块 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将后端所有大模型交互封装到 `internal/infrastructure/llm` 通用模块，文本对话使用 eino 框架，图片生成使用原生 HTTP 封装，对外完全隐藏 eino 细节。

**Architecture:** 在 `internal/infrastructure/llm` 定义 `TextClient` 和 `ImageClient` 两个接口，分别由 `textClient`（内部基于 eino openai 组件）和 `imageClient`（内部基于 net/http）实现。Handler 和 Job 通过接口注入，配置统一从 `config.yaml` 的 `llm` 区块读取。

**Tech Stack:** Go 1.22, CloudWeGo Eino (`github.com/cloudwego/eino`), `eino-ext/components/model/openai`, Gin, testify

---

## File Structure

| 文件 | 职责 |
|---|---|
| `internal/infrastructure/llm/interfaces.go` | 公共接口与类型：`TextClient`, `ImageClient`, `Message`, `Part`, `StreamReader`, `Chunk`, `TextConfig`, `ImageConfig` |
| `internal/infrastructure/llm/convert.go` | `llm.Message` ↔ `schema.Message`（eino 内部类型）转换 |
| `internal/infrastructure/llm/convert_test.go` | 转换逻辑单元测试 |
| `internal/infrastructure/llm/text_client.go` | `textClient` 实现，内部调用 eino `openai.ChatModel` |
| `internal/infrastructure/llm/text_client_test.go` | text client 单元测试（httptest mock eino 底层 HTTP） |
| `internal/infrastructure/llm/image_client.go` | `imageClient` 实现，直接 HTTP 调用文生图 API |
| `internal/infrastructure/llm/image_client_test.go` | image client 单元测试（httptest mock） |
| `internal/config/config.go` | 新增 `LLMConfig`/`TextConfig`/`ImageConfig`，保留 DMX fallback |
| `backend/config.yaml` | 新增 `llm` 配置区块 |
| `internal/http/handlers/face_reading_handler.go` | 改为注入 `llm.TextClient`，使用 `ChatStream` |
| `internal/http/handlers/face_reading_handler_test.go` | 改为 mock `llm.TextClient`，验证 SSE 输出 |
| `internal/worker/jobs/generation_job.go` | `ModelClient` 接口替换为 `llm.ImageClient` |
| `internal/worker/jobs/generation_job_test.go` | 适配 `llm.ImageClient`（`captureModelClient` 签名天然匹配） |
| `internal/http/router.go` | 接收 `llm.TextClient` 与 `llm.ImageClient` 并注入 Handler |
| `cmd/api/main.go` | 初始化 `llm.NewTextClient` 传入 Router |
| `cmd/worker/main.go` | 初始化 `llm.NewImageClient` 传入 `GenerationJob` |
| `CLAUDE.md` | 新增「大模型交互技术选型」章节 |

---

### Task 1: Define LLM module interfaces

**Files:**
- Create: `internal/infrastructure/llm/interfaces.go`

- [ ] **Step 1: Write interfaces.go**

```go
package llm

import (
	"context"
	"time"
)

type PartType string

const (
	PartTypeText  PartType = "text"
	PartTypeImage PartType = "image"
)

type Part struct {
	Type    PartType
	Content string
}

type Message struct {
	Role  string
	Parts []Part
}

type Chunk struct {
	Content string
}

// StreamReader 流式读取模型输出
// Recv 返回 io.EOF 表示流正常结束
// 调用方必须在使用完成后调用 Close
type StreamReader interface {
	Recv() (Chunk, error)
	Close()
}

// TextClient 文本/多模态大模型客户端
type TextClient interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	ChatStream(ctx context.Context, messages []Message) (StreamReader, error)
}

// ImageClient 文生图大模型客户端
type ImageClient interface {
	Generate(ctx context.Context, prompt string) (imageURL string, err error)
}

type TextConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

type ImageConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/infrastructure/llm/interfaces.go
git commit -m "feat(llm): define public interfaces and types

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 2: Implement message conversion + tests

**Files:**
- Create: `internal/infrastructure/llm/convert.go`
- Create: `internal/infrastructure/llm/convert_test.go`

- [ ] **Step 1: Write convert.go**

```go
package llm

import (
	"fmt"

	"github.com/cloudwego/eino/schema"
)

func toSchemaMessages(messages []Message) ([]*schema.Message, error) {
	result := make([]*schema.Message, 0, len(messages))
	for _, msg := range messages {
		m := &schema.Message{
			Role: schema.RoleType(msg.Role),
		}
		if len(msg.Parts) == 0 {
			result = append(result, m)
			continue
		}
		if len(msg.Parts) == 1 && msg.Parts[0].Type == PartTypeText {
			m.Content = msg.Parts[0].Content
			result = append(result, m)
			continue
		}
		parts := make([]schema.MessageInputPart, 0, len(msg.Parts))
		for _, part := range msg.Parts {
			switch part.Type {
			case PartTypeText:
				parts = append(parts, schema.MessageInputPart{
					Type: schema.ChatMessagePartTypeText,
					Text: part.Content,
				})
			case PartTypeImage:
				parts = append(parts, schema.MessageInputPart{
					Type: schema.ChatMessagePartTypeImageURL,
					Image: &schema.MessageInputImage{
						MessagePartCommon: schema.MessagePartCommon{
							URL: &part.Content,
						},
					},
				})
			default:
				return nil, fmt.Errorf("llm: unsupported part type %q", part.Type)
			}
		}
		m.UserInputMultiContent = parts
		result = append(result, m)
	}
	return result, nil
}
```

- [ ] **Step 2: Write convert_test.go**

```go
package llm

import (
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSchemaMessages_TextOnly(t *testing.T) {
	msgs := []Message{
		{Role: "user", Parts: []Part{{Type: PartTypeText, Content: "hello"}}},
	}
	result, err := toSchemaMessages(msgs)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, schema.User, result[0].Role)
	assert.Equal(t, "hello", result[0].Content)
	assert.Empty(t, result[0].UserInputMultiContent)
}

func TestToSchemaMessages_Multimodal(t *testing.T) {
	msgs := []Message{
		{
			Role: "user",
			Parts: []Part{
				{Type: PartTypeText, Content: "analyze this"},
				{Type: PartTypeImage, Content: "data:image/png;base64,abc"},
			},
		},
	}
	result, err := toSchemaMessages(msgs)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, schema.User, result[0].Role)
	assert.Empty(t, result[0].Content)
	require.Len(t, result[0].UserInputMultiContent, 2)
	assert.Equal(t, schema.ChatMessagePartTypeText, result[0].UserInputMultiContent[0].Type)
	assert.Equal(t, "analyze this", result[0].UserInputMultiContent[0].Text)
	assert.Equal(t, schema.ChatMessagePartTypeImageURL, result[0].UserInputMultiContent[1].Type)
	assert.Equal(t, "data:image/png;base64,abc", *result[0].UserInputMultiContent[1].Image.URL)
}

func TestToSchemaMessages_UnsupportedPartType(t *testing.T) {
	msgs := []Message{
		{Role: "user", Parts: []Part{{Type: PartType("audio"), Content: "x"}}},
	}
	_, err := toSchemaMessages(msgs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported part type")
}
```

- [ ] **Step 3: Run tests**

```bash
cd /root/project/image-play/backend
go test ./internal/infrastructure/llm/... -v
```
Expected: `PASS` for `TestToSchemaMessages_TextOnly`, `TestToSchemaMessages_Multimodal`, `TestToSchemaMessages_UnsupportedPartType`.

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/llm/convert.go internal/infrastructure/llm/convert_test.go
git commit -m "feat(llm): add message conversion and tests

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 3: Implement text client + tests

**Files:**
- Create: `internal/infrastructure/llm/text_client.go`
- Create: `internal/infrastructure/llm/text_client_test.go`

- [ ] **Step 1: Write text_client.go**

```go
package llm

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type textClient struct {
	chatModel model.ChatModel
}

func NewTextClient(cfg TextConfig) (TextClient, error) {
	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,
		Timeout: cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("llm: create text client: %w", err)
	}
	return &textClient{chatModel: chatModel}, nil
}

func (c *textClient) Chat(ctx context.Context, messages []Message) (string, error) {
	inputs, err := toSchemaMessages(messages)
	if err != nil {
		return "", err
	}
	resp, err := c.chatModel.Generate(ctx, inputs)
	if err != nil {
		return "", fmt.Errorf("llm: chat generate failed: %w", err)
	}
	return resp.Content, nil
}

func (c *textClient) ChatStream(ctx context.Context, messages []Message) (StreamReader, error) {
	inputs, err := toSchemaMessages(messages)
	if err != nil {
		return nil, err
	}
	stream, err := c.chatModel.Stream(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("llm: chat stream failed: %w", err)
	}
	return &textStreamReader{stream: stream}, nil
}

type textStreamReader struct {
	stream *schema.StreamReader[*schema.Message]
}

func (r *textStreamReader) Recv() (Chunk, error) {
	msg, err := r.stream.Recv()
	if err != nil {
		return Chunk{}, err
	}
	return Chunk{Content: msg.Content}, nil
}

func (r *textStreamReader) Close() {
	r.stream.Close()
}
```

- [ ] **Step 2: Write text_client_test.go**

```go
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
```

- [ ] **Step 3: Run tests**

```bash
cd /root/project/image-play/backend
go test ./internal/infrastructure/llm/... -run TestTextClient -v
```
Expected: `PASS` for `TestTextClient_Chat` and `TestTextClient_ChatStream`.

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/llm/text_client.go internal/infrastructure/llm/text_client_test.go
git commit -m "feat(llm): implement text client with eino

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 4: Implement image client + tests

**Files:**
- Create: `internal/infrastructure/llm/image_client.go`
- Create: `internal/infrastructure/llm/image_client_test.go`

- [ ] **Step 1: Write image_client.go**

```go
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
```

- [ ] **Step 2: Write image_client_test.go**

```go
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
```

- [ ] **Step 3: Run tests**

```bash
cd /root/project/image-play/backend
go test ./internal/infrastructure/llm/... -run TestImageClient -v
```
Expected: `PASS` for `TestImageClient_Generate` and `TestImageClient_Generate_APIError`.

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/llm/image_client.go internal/infrastructure/llm/image_client_test.go
git commit -m "feat(llm): implement image client

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 5: Update configuration

**Files:**
- Modify: `internal/config/config.go`
- Modify: `backend/config.yaml`

- [ ] **Step 1: Modify config.go**

Add `LLMConfig`, `TextConfig`, `ImageConfig` types, update `Config` struct, and add fallback logic.

Replace the `Config` struct definition with:

```go
type Config struct {
	AppEnv          string `yaml:"app_env"`
	Port            string `yaml:"port"`
	DatabaseURL     string `yaml:"database_url"`
	JWTSecret       string `yaml:"jwt_secret"`
	WechatAppID        string `yaml:"wechat_app_id"`
	WechatAppSecret    string `yaml:"wechat_app_secret"`
	OSSEndpoint        string `yaml:"oss_endpoint"`
	OSSBucket          string `yaml:"oss_bucket"`
	OSSAccessKeyID     string `yaml:"oss_access_key_id"`
	OSSAccessKeySecret string `yaml:"oss_access_key_secret"`
	DMXAPIKey          string `yaml:"dmx_api_key"`
	DMXAPIBaseURL      string `yaml:"dmx_api_base_url"`
	DMXModel           string `yaml:"dmx_model"`
	LLM                LLMConfig `yaml:"llm"`
}

type LLMConfig struct {
	Text  TextConfig  `yaml:"text"`
	Image ImageConfig `yaml:"image"`
}

type TextConfig struct {
	APIKey  string        `yaml:"api_key"`
	BaseURL string        `yaml:"base_url"`
	Model   string        `yaml:"model"`
	Timeout time.Duration `yaml:"timeout"`
}

type ImageConfig struct {
	APIKey  string        `yaml:"api_key"`
	BaseURL string        `yaml:"base_url"`
	Model   string        `yaml:"model"`
	Timeout time.Duration `yaml:"timeout"`
}
```

In `Load()`, after loading the config file and env overrides, add this block before `return &cfg`:

```go
	// Backward compatibility: if llm.text is empty, fall back to dmx_* fields
	if cfg.LLM.Text.APIKey == "" {
		cfg.LLM.Text.APIKey = cfg.DMXAPIKey
	}
	if cfg.LLM.Text.BaseURL == "" {
		cfg.LLM.Text.BaseURL = cfg.DMXAPIBaseURL
	}
	if cfg.LLM.Text.Model == "" {
		cfg.LLM.Text.Model = cfg.DMXModel
	}
	if cfg.LLM.Text.Timeout == 0 {
		cfg.LLM.Text.Timeout = 300 * time.Second
	}
	if cfg.LLM.Image.Timeout == 0 {
		cfg.LLM.Image.Timeout = 60 * time.Second
	}
```

Also add `time` import if missing.

In `loadFromEnv()`, update to include `LLM` fallback from env vars:

```go
func loadFromEnv() *Config {
	return &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		WechatAppID:        getEnv("WECHAT_APP_ID", ""),
		WechatAppSecret:    getEnv("WECHAT_APP_SECRET", ""),
		OSSEndpoint:        getEnv("OSS_ENDPOINT", ""),
		OSSBucket:          getEnv("OSS_BUCKET", ""),
		OSSAccessKeyID:     getEnv("OSS_ACCESS_KEY_ID", ""),
		OSSAccessKeySecret: getEnv("OSS_ACCESS_KEY_SECRET", ""),
		DMXAPIKey:          getEnv("DMX_API_KEY", ""),
		DMXAPIBaseURL:      getEnv("DMX_API_BASE_URL", "https://api.moonshot.cn/v1"),
		DMXModel:           getEnv("DMX_MODEL", "kimi-k2.6"),
		LLM: LLMConfig{
			Text: TextConfig{
				APIKey:  getEnv("LLM_TEXT_API_KEY", ""),
				BaseURL: getEnv("LLM_TEXT_BASE_URL", ""),
				Model:   getEnv("LLM_TEXT_MODEL", ""),
			},
			Image: ImageConfig{
				APIKey:  getEnv("LLM_IMAGE_API_KEY", ""),
				BaseURL: getEnv("LLM_IMAGE_BASE_URL", ""),
				Model:   getEnv("LLM_IMAGE_MODEL", ""),
			},
		},
	}
}
```

And add env overrides for LLM fields in `Load()` after the existing env var overrides:

```go
	if v := os.Getenv("LLM_TEXT_API_KEY"); v != "" {
		cfg.LLM.Text.APIKey = v
	}
	if v := os.Getenv("LLM_TEXT_BASE_URL"); v != "" {
		cfg.LLM.Text.BaseURL = v
	}
	if v := os.Getenv("LLM_TEXT_MODEL"); v != "" {
		cfg.LLM.Text.Model = v
	}
	if v := os.Getenv("LLM_IMAGE_API_KEY"); v != "" {
		cfg.LLM.Image.APIKey = v
	}
	if v := os.Getenv("LLM_IMAGE_BASE_URL"); v != "" {
		cfg.LLM.Image.BaseURL = v
	}
	if v := os.Getenv("LLM_IMAGE_MODEL"); v != "" {
		cfg.LLM.Image.Model = v
	}
```

- [ ] **Step 2: Modify config.yaml**

Append the `llm` block to `backend/config.yaml`:

```yaml
llm:
  text:
    api_key: "sk-2mRuFylp26J7PNkyI3He9MWHqI4maxbudu5lJzntZEQOhvQ3"
    base_url: "https://api.moonshot.cn/v1"
    model: "kimi-k2.6"
    timeout: 300
  image:
    api_key: ""
    base_url: ""
    model: ""
    timeout: 60
```

- [ ] **Step 3: Commit**

```bash
git add internal/config/config.go backend/config.yaml
git commit -m "feat(config): add llm configuration with backward compatibility

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 6: Refactor face reading handler

**Files:**
- Modify: `internal/http/handlers/face_reading_handler.go`
- Modify: `internal/http/handlers/face_reading_handler_test.go`

- [ ] **Step 1: Rewrite face_reading_handler.go**

Replace the entire file content:

```go
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/infrastructure/llm"
)

const faceReadingPrompt = `你是一位精通面相学与命理学的资深相师。请基于中国传统面相学、周易命理学及现代心理学的交叉视角，对图片中人物进行深度解析。分析维度必须包括以下方面：

1. 面相格局与五行属性：分析脸型对应的五行属性及整体格局高低。
2. 五官详解：眉毛、眼睛、鼻子、嘴巴、耳朵分别对应的宫位与运势。
3. 三停比例：上停、中停、下停的均衡度与运势走向。
4. 十二宫位速览：命宫、财帛宫、夫妻宫、疾厄宫等关键宫位。
5. 性格与气质：内在性格、情绪模式与待人接物风格。
6. 情感运势：感情观、桃花运、婚姻稳定性。
7. 事业财运：事业发展潜力、财富积累能力与适合方向。
8. 健康提示：面色、眼神等反映的体质倾向。

要求：请用专业且通俗的语言输出，分点清晰，有理有据。分析仅供参考娱乐，请保持客观理性，避免绝对化断言。`

type FaceReadingRequest struct {
	ImageBase64 string `json:"image_base64" binding:"required"`
}

func FaceReadingHandler(textClient llm.TextClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req FaceReadingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image_base64 is required"})
			return
		}

		const maxBase64Len = 7 * 1024 * 1024
		if len(req.ImageBase64) > maxBase64Len {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image too large"})
			return
		}

		messages := []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Type: llm.PartTypeText, Content: faceReadingPrompt},
					{Type: llm.PartTypeImage, Content: req.ImageBase64},
				},
			},
		}

		reader, err := textClient.ChatStream(c.Request.Context(), messages)
		if err != nil {
			fmt.Printf("[face-reading] chat stream error: %v\n", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "model unavailable"})
			return
		}
		defer reader.Close()

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Writer.WriteHeader(http.StatusOK)

		for {
			chunk, err := reader.Recv()
			if err == io.EOF {
				c.Writer.WriteString("data: [DONE]\n\n")
				c.Writer.Flush()
				break
			}
			if err != nil {
				fmt.Printf("[face-reading] recv error: %v\n", err)
				break
			}
			if chunk.Content != "" {
				out, _ := json.Marshal(map[string]string{"chunk": chunk.Content})
				c.Writer.WriteString("data: " + string(out) + "\n\n")
				c.Writer.Flush()
			}
		}
	}
}
```

- [ ] **Step 2: Rewrite face_reading_handler_test.go**

Replace the entire file content:

```go
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
```

- [ ] **Step 3: Run tests**

```bash
cd /root/project/image-play/backend
go test ./internal/http/handlers/... -run TestFaceReadingHandler -v
```
Expected: `PASS` for all three tests.

- [ ] **Step 4: Commit**

```bash
git add internal/http/handlers/face_reading_handler.go internal/http/handlers/face_reading_handler_test.go
git commit -m "refactor(face-reading): use llm.TextClient for streaming

Replace raw HTTP with llm module ChatStream.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 7: Refactor generation job

**Files:**
- Modify: `internal/worker/jobs/generation_job.go`
- Modify: `internal/worker/jobs/generation_job_test.go`

- [ ] **Step 1: Modify generation_job.go**

Replace the `ModelClient` interface and `MockModelClient` with usage of `llm.ImageClient`.

Remove from the file:
```go
type ModelClient interface {
	Generate(ctx context.Context, prompt string) (imageURL string, err error)
}

type MockModelClient struct{}

func (m *MockModelClient) Generate(ctx context.Context, prompt string) (string, error) {
	// Simulate model generation latency
	select {
	case <-time.After(100 * time.Millisecond):
		return fmt.Sprintf("https://mock-cdn.example.com/result/%d.png", time.Now().Unix()), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
```

Add import for `image-play/internal/infrastructure/llm`.

Change the `GenerationJob` struct:

```go
type GenerationJob struct {
	imageClient    llm.ImageClient
	auditClient    AuditClient
	generationRepo generation.Repository
	templateRepo   generation.TemplateLookup
	billingSvc     *billing.Service
}
```

Change constructor:

```go
func NewGenerationJob(repo generation.Repository, templateRepo generation.TemplateLookup, imageClient llm.ImageClient, audit AuditClient, billingSvc *billing.Service) *GenerationJob {
	if templateRepo == nil {
		panic("template lookup is required")
	}
	if imageClient == nil {
		panic("image client is required")
	}
	if audit == nil {
		audit = &MockAuditClient{}
	}
	return &GenerationJob{
		imageClient:    imageClient,
		auditClient:    audit,
		generationRepo: repo,
		templateRepo:   templateRepo,
		billingSvc:     billingSvc,
	}
}
```

Change the call in `Execute` from `j.modelClient.Generate` to `j.imageClient.Generate`.

- [ ] **Step 2: Modify generation_job_test.go**

The existing `captureModelClient` already implements `llm.ImageClient` because its `Generate` method signature matches exactly. We just need to update the import and the constructor call.

Add import:
```go
	"image-play/internal/infrastructure/llm"
```

In `TestExecuteBuildsPromptFromTemplatePreset`, update the constructor call:

```go
	job := NewGenerationJob(repo, templateLoader, model, &passAuditClient{}, nil)
```

Do the same for `TestExecuteRejectsTemplateWithInvalidPromptPreset`.

In `TestNewGenerationJobRequiresTemplateLookup`, the test currently passes `nil` for model and audit. Since we now panic when imageClient is nil, update to:

```go
func TestNewGenerationJobRequiresTemplateLookup(t *testing.T) {
	require.PanicsWithValue(t, "template lookup is required", func() {
		NewGenerationJob(&stubGenerationRepo{}, nil, &captureModelClient{}, nil, nil)
	})
}
```

Also add a test for image client nil panic:

```go
func TestNewGenerationJobRequiresImageClient(t *testing.T) {
	require.PanicsWithValue(t, "image client is required", func() {
		NewGenerationJob(&stubGenerationRepo{}, &stubTemplateLoader{}, nil, nil, nil)
	})
}
```

- [ ] **Step 3: Run tests**

```bash
cd /root/project/image-play/backend
go test ./internal/worker/jobs/... -v
```
Expected: `PASS` for all tests.

- [ ] **Step 4: Commit**

```bash
git add internal/worker/jobs/generation_job.go internal/worker/jobs/generation_job_test.go
git commit -m "refactor(worker): replace ModelClient with llm.ImageClient

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 8: Wire up router and main entry points

**Files:**
- Modify: `internal/http/router.go`
- Modify: `cmd/api/main.go`
- Modify: `cmd/worker/main.go`

- [ ] **Step 1: Modify router.go**

Update `NewRouter` signature and add `llm` import:

```go
import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"image-play/internal/config"
	"image-play/internal/domain/assets"
	"image-play/internal/domain/billing"
	"image-play/internal/domain/generation"
	"image-play/internal/domain/tracking"
	"image-play/internal/domain/user"
	"image-play/internal/http/handlers"
	"image-play/internal/http/middleware"
	"image-play/internal/infrastructure/llm"
	"image-play/internal/infrastructure/storage"
	"image-play/internal/infrastructure/wechat"
	"image-play/internal/repository/postgres"
)
```

Change function signature:

```go
func NewRouter(db *sql.DB, cfg *config.Config, wxClient *wechat.Client, signer storage.Signer, textClient llm.TextClient) *gin.Engine {
```

Change the face-reading route registration:

```go
	authorized.POST("/face-reading", handlers.FaceReadingHandler(textClient))
```

- [ ] **Step 2: Modify cmd/api/main.go**

Add import:
```go
	"image-play/internal/infrastructure/llm"
```

Before creating the router, initialize the text client:

```go
	textClient, err := llm.NewTextClient(llm.TextConfig{
		APIKey:  cfg.LLM.Text.APIKey,
		BaseURL: cfg.LLM.Text.BaseURL,
		Model:   cfg.LLM.Text.Model,
		Timeout: cfg.LLM.Text.Timeout,
	})
	if err != nil {
		log.Fatalf("text client init failed: %v", err)
	}

	r := http.NewRouter(db, cfg, wxClient, signer, textClient)
```

- [ ] **Step 3: Modify cmd/worker/main.go**

Add import:
```go
	"image-play/internal/infrastructure/llm"
```

After loading config and before creating the job, initialize the image client:

```go
	imageClient, err := llm.NewImageClient(llm.ImageConfig{
		APIKey:  cfg.LLM.Image.APIKey,
		BaseURL: cfg.LLM.Image.BaseURL,
		Model:   cfg.LLM.Image.Model,
		Timeout: cfg.LLM.Image.Timeout,
	})
	if err != nil {
		log.Fatalf("image client init failed: %v", err)
	}

	job := jobs.NewGenerationJob(repo, templateRepo, imageClient, nil, billingSvc)
```

- [ ] **Step 4: Build verification**

```bash
cd /root/project/image-play/backend
go build ./cmd/api
go build ./cmd/worker
```
Expected: both build successfully with no errors.

- [ ] **Step 5: Commit**

```bash
git add internal/http/router.go cmd/api/main.go cmd/worker/main.go
git commit -m "feat(wiring): inject llm clients into router and workers

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 9: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Append LLM tech stack section**

Add the following section near the top of `CLAUDE.md`, right after the `## 技术栈` section:

```markdown
### 大模型交互技术选型

- **LLM 应用框架**: [CloudWeGo Eino](https://github.com/cloudwego/eino) — 构建大语言模型应用的 Go 框架，提供统一的组件抽象、流式处理、回调机制。
- **文本模型组件**: `eino-ext/components/model/openai` — 兼容 OpenAI API 格式的 ChatModel 实现。Moonshot 等国内主流模型均兼容此格式。
- **图片生成**: 原生 HTTP 封装 — eino 生态暂无标准文生图组件，直接封装更可控。
- **模块位置**: `backend/internal/infrastructure/llm/`
- **接口设计**: 项目自定义 `TextClient` / `ImageClient` 接口，完全封装 eino 细节，调用方零依赖。
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs(CLAUDE): document LLM technology choices

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 10: Full test suite verification

- [ ] **Step 1: Run all backend tests**

```bash
cd /root/project/image-play/backend
go test ./... -count=1
```
Expected: all tests pass.

- [ ] **Step 2: Run backend build**

```bash
cd /root/project/image-play/backend
go build ./cmd/api
go build ./cmd/worker
```
Expected: clean builds.

- [ ] **Step 3: Final commit (if any remaining changes)**

```bash
git diff --stat
git add -A
git commit -m "feat(llm): complete generic LLM interaction module

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Self-Review Checklist

**1. Spec coverage:**
- [x] 定义公共接口与类型 → Task 1
- [x] 消息转换逻辑 → Task 2
- [x] 文本客户端（eino 实现） → Task 3
- [x] 图片客户端（HTTP 实现） → Task 4
- [x] 配置与向下兼容 → Task 5
- [x] 面相分析 Handler 重构 → Task 6
- [x] 图片生成 Job 重构 → Task 7
- [x] Router 与 main 接入 → Task 8
- [x] CLAUDE.md 文档更新 → Task 9
- [x] 全量测试验证 → Task 10

**2. Placeholder scan:** 无 TBD、TODO、"implement later"。所有步骤均包含完整代码。

**3. Type consistency:**
- `TextClient.ChatStream` 返回 `(StreamReader, error)`，与 `textStreamReader` 实现一致。
- `ImageClient.Generate` 签名与旧 `ModelClient.Generate` 完全一致，测试无需改动 mock 结构。
- `FaceReadingHandler` 参数由 `apiKey, baseURL, model string` 统一改为 `llm.TextClient`。
- `NewRouter` 新增 `textClient llm.TextClient` 参数。

**4. Scope check:** 聚焦在 LLM 模块封装，不引入无关重构（如 AuditClient 保持不变，billing 逻辑不变）。
