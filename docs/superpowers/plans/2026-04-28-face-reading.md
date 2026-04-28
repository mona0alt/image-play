# 面相分析功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 image-play 小程序中集成 DMXAPI 面相分析功能，用户从首页进入，上传照片后由后端代理调用 DMXAPI 并返回分析结果。

**Architecture:** 后端新增 `/api/face-reading` 接口（JWT 保护），使用标准 `net/http` 代理请求到 DMXAPI；前端新增「面相分析」页面，支持拍照/选图、预览、提交和结果复制。

**Tech Stack:** Go 1.22, Gin, Vue 3, uni-app, TypeScript

---

## 文件结构

| 文件 | 操作 | 职责 |
|---|---|---|
| `backend/internal/config/config.go` | 修改 | 新增 DMXAPI 配置字段 |
| `backend/internal/config/config_test.go` | 创建 | 配置加载单元测试 |
| `backend/config.yaml` | 修改 | 添加配置示例 |
| `backend/internal/http/handlers/face_reading_handler.go` | 创建 | 面相分析接口 handler |
| `backend/internal/http/handlers/face_reading_handler_test.go` | 创建 | handler 单元测试 |
| `backend/internal/http/router.go` | 修改 | 注册新路由 |
| `frontend/src/services/api.ts` | 修改 | 新增 `faceReading` API 调用 |
| `frontend/src/pages/face-reading/index.vue` | 创建 | 面相分析页面 |
| `frontend/src/pages.json` | 修改 | 注册新页面 |
| `frontend/src/pages/home/index.vue` | 修改 | 添加功能入口卡片 |

---

### Task 1: 扩展后端配置

**Files:**
- Modify: `backend/internal/config/config.go`
- Create: `backend/internal/config/config_test.go`
- Modify: `backend/config.yaml`

- [ ] **Step 1: 写配置加载测试**

创建 `backend/internal/config/config_test.go`：

```go
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
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /root/project/image-play/backend
go test ./internal/config/... -v
```

Expected: FAIL — `DMXAPIKey` 等字段未定义

- [ ] **Step 3: 修改 `config.go` 添加字段**

在 `Config` struct 中新增：

```go
	type Config struct {
		// ... existing fields ...
		DMXAPIKey     string `yaml:"dmx_api_key"`
		DMXAPIBaseURL string `yaml:"dmx_api_base_url"`
		DMXModel      string `yaml:"dmx_model"`
	}
```

在 `Load()` 函数的环境变量覆盖段新增：

```go
	if v := os.Getenv("DMX_API_KEY"); v != "" {
		cfg.DMXAPIKey = v
	}
	if v := os.Getenv("DMX_API_BASE_URL"); v != "" {
		cfg.DMXAPIBaseURL = v
	}
	if v := os.Getenv("DMX_MODEL"); v != "" {
		cfg.DMXModel = v
	}
```

在 `loadFromEnv()` 中新增：

```go
		DMXAPIKey:     getEnv("DMX_API_KEY", ""),
		DMXAPIBaseURL: getEnv("DMX_API_BASE_URL", "https://api.moonshot.cn/v1"),
		DMXModel:      getEnv("DMX_MODEL", "kimi-k2.6"),
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /root/project/image-play/backend
go test ./internal/config/... -v
```

Expected: PASS

- [ ] **Step 5: 修改 `backend/config.yaml` 添加示例**

在文件末尾新增：

```yaml
# DMXAPI 配置（面相分析）
dmx_api_key: ""
dmx_api_base_url: "https://api.moonshot.cn/v1"
dmx_model: "kimi-k2.6"
```

- [ ] **Step 6: 提交**

```bash
git add backend/internal/config/config.go backend/internal/config/config_test.go backend/config.yaml
git commit -m "$(cat <<'EOF'
feat(config): add DMXAPI configuration for face reading

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: 实现面相分析 Handler

**Files:**
- Create: `backend/internal/http/handlers/face_reading_handler.go`
- Create: `backend/internal/http/handlers/face_reading_handler_test.go`

- [ ] **Step 1: 写 handler 单元测试**

创建 `backend/internal/http/handlers/face_reading_handler_test.go`：

```go
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

	// 模拟 DMXAPI 服务器
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
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /root/project/image-play/backend
go test ./internal/http/handlers/... -run TestFaceReading -v
```

Expected: FAIL — `FaceReadingHandler` 未定义

- [ ] **Step 3: 实现 handler**

创建 `backend/internal/http/handlers/face_reading_handler.go`：

```go
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

type FaceReadingResponse struct {
	Result string `json:"result"`
}

type dmxMessage struct {
	Role    string `json:"role"`
	Content []dmxContent `json:"content"`
}

type dmxContent struct {
	Type     string            `json:"type"`
	Text     string            `json:"text,omitempty"`
	ImageURL map[string]string `json:"image_url,omitempty"`
}

type dmxRequest struct {
	Model    string       `json:"model"`
	Messages []dmxMessage `json:"messages"`
}

type dmxResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func FaceReadingHandler(apiKey, baseURL, model string) gin.HandlerFunc {
	client := &http.Client{Timeout: 60 * time.Second}

	return func(c *gin.Context) {
		var req FaceReadingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image_base64 is required"})
			return
		}

		// 限制图片大小（base64 长度约等于原始大小的 4/3）
		const maxBase64Len = 7 * 1024 * 1024 // 约 5MB 原始图片
		if len(req.ImageBase64) > maxBase64Len {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image too large"})
			return
		}

		dmxReq := dmxRequest{
			Model: model,
			Messages: []dmxMessage{
				{
					Role: "user",
					Content: []dmxContent{
						{Type: "text", Text: faceReadingPrompt},
						{Type: "image_url", ImageURL: map[string]string{"url": req.ImageBase64}},
					},
				},
			},
		}

		bodyBytes, err := json.Marshal(dmxReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
			return
		}

		httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "dmx api unreachable"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(gin.DefaultWriter, "DMXAPI error: status=%d body=%s\n", resp.StatusCode, string(body))
			c.JSON(http.StatusBadGateway, gin.H{"error": "dmx api error"})
			return
		}

		var dmxResp dmxResponse
		if err := json.NewDecoder(resp.Body).Decode(&dmxResp); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "invalid dmx response"})
			return
		}

		if len(dmxResp.Choices) == 0 {
			c.JSON(http.StatusBadGateway, gin.H{"error": "empty dmx response"})
			return
		}

		c.JSON(http.StatusOK, FaceReadingResponse{Result: dmxResp.Choices[0].Message.Content})
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /root/project/image-play/backend
go test ./internal/http/handlers/... -run TestFaceReading -v
```

Expected: PASS（3 个测试全部通过）

- [ ] **Step 5: 提交**

```bash
git add backend/internal/http/handlers/face_reading_handler.go backend/internal/http/handlers/face_reading_handler_test.go
git commit -m "$(cat <<'EOF'
feat(face-reading): add backend handler with DMXAPI proxy

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: 注册路由

**Files:**
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: 修改 router.go 注册接口**

在 `authorized` 组内（JWT 保护下），新增路由：

```go
	authorized.POST("/face-reading", handlers.FaceReadingHandler(cfg.DMXAPIKey, cfg.DMXAPIBaseURL, cfg.DMXModel))
```

注意：`NewRouter` 函数签名需要传入 `cfg *config.Config`（或单独传入 dmx 参数）。当前 `NewRouter` 接收的是分散参数，为最小改动，直接修改函数签名添加 dmx 参数，或传入 cfg。

修改 `NewRouter` 签名和调用处：

```go
func NewRouter(db *sql.DB, cfg *config.Config, wxClient *wechat.Client, signer storage.Signer) *gin.Engine {
```

然后在 `authorized` 组内添加：

```go
	authorized.POST("/face-reading", handlers.FaceReadingHandler(cfg.DMXAPIKey, cfg.DMXAPIBaseURL, cfg.DMXModel))
```

同时修改 `cmd/api/main.go` 中的 `NewRouter` 调用，将 `jwtSecret` 替换为 `cfg`：

```go
	router := http.NewRouter(db, cfg, wxClient, signer)
```

- [ ] **Step 2: 编译确认通过**

```bash
cd /root/project/image-play/backend
go build ./cmd/api
```

Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add backend/internal/http/router.go backend/cmd/api/main.go
git commit -m "$(cat <<'EOF'
feat(router): register face-reading endpoint

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: 前端 API 服务

**Files:**
- Modify: `frontend/src/services/api.ts`

- [ ] **Step 1: 在 `api.ts` 中添加面相分析请求函数**

在文件末尾新增：

```typescript
export function faceReading(imageBase64: string) {
  return request<{ result: string }>({
    url: '/api/face-reading',
    method: 'POST',
    data: { image_base64: imageBase64 },
    headers: { 'Content-Type': 'application/json' },
  })
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/services/api.ts
git commit -m "$(cat <<'EOF'
feat(api): add faceReading service function

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: 前端面相分析页面

**Files:**
- Create: `frontend/src/pages/face-reading/index.vue`

- [ ] **Step 1: 创建页面组件**

创建 `frontend/src/pages/face-reading/index.vue`：

```vue
<script setup lang="ts">
import { ref } from 'vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import { faceReading } from '../../services/api'

const imageBase64 = ref('')
const result = ref('')
const loading = ref(false)
const error = ref('')

function chooseImage() {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (res: any) => {
      const tempPath = res.tempFilePaths[0] as string
      const fs = uni.getFileSystemManager()
      fs.readFile({
        filePath: tempPath,
        encoding: 'base64',
        success: (readRes: any) => {
          const ext = tempPath.split('.').pop()?.toLowerCase() || 'jpeg'
          const mime = ext === 'png' ? 'image/png' : ext === 'gif' ? 'image/gif' : ext === 'webp' ? 'image/webp' : 'image/jpeg'
          imageBase64.value = `data:${mime};base64,${readRes.data}`
          result.value = ''
          error.value = ''
        },
        fail: () => {
          uni.showToast({ title: '读取图片失败', icon: 'none' })
        },
      })
    },
    fail: () => {
      uni.showToast({ title: '选择图片失败', icon: 'none' })
    },
  })
}

function reset() {
  imageBase64.value = ''
  result.value = ''
  error.value = ''
}

async function submit() {
  if (!imageBase64.value) {
    uni.showToast({ title: '请先选择一张照片', icon: 'none' })
    return
  }
  if (imageBase64.value.length > 7 * 1024 * 1024) {
    uni.showToast({ title: '图片过大，请选择较小的图片', icon: 'none' })
    return
  }

  loading.value = true
  error.value = ''
  try {
    const res = await faceReading(imageBase64.value)
    result.value = res.result
  } catch (e: any) {
    error.value = '分析服务暂时繁忙，请稍后重试'
  } finally {
    loading.value = false
  }
}

function copyResult() {
  if (!result.value) return
  uni.setClipboardData({
    data: result.value,
    success: () => {
      uni.showToast({ title: '已复制', icon: 'success' })
    },
  })
}
</script>

<template>
  <GalleryPageShell active-tab="gallery" title="面相分析">
    <view class="face-page">
      <view class="face-page__hero">
        <text class="face-page__hero-title">面相分析</text>
        <text class="face-page__hero-desc">上传一张正面清晰照片，AI 将基于传统面相学进行解析</text>
      </view>

      <view v-if="!imageBase64" class="face-page__upload-card" @click="chooseImage">
        <text class="face-page__upload-icon">+</text>
        <text class="face-page__upload-label">选择照片</text>
        <text class="face-page__upload-hint">支持拍照或从相册选择</text>
      </view>

      <view v-else class="face-page__preview-card">
        <image class="face-page__preview-image" :src="imageBase64" mode="aspectFit" />
        <view class="face-page__preview-actions">
          <button class="face-page__btn-secondary" @click="reset">重新选择</button>
          <button class="face-page__btn-primary" :disabled="loading" @click="submit">
            {{ loading ? '正在分析…' : '开始分析' }}
          </button>
        </view>
      </view>

      <view v-if="loading" class="face-page__loading">
        <text class="face-page__loading-text">正在分析面相，请稍候…</text>
      </view>

      <view v-if="error" class="face-page__error">
        <text>{{ error }}</text>
      </view>

      <view v-if="result" class="face-page__result-card">
        <text class="face-page__result-title">分析结果</text>
        <text class="face-page__result-body">{{ result }}</text>
        <button class="face-page__btn-primary" @click="copyResult">复制结果</button>
      </view>
    </view>
  </GalleryPageShell>
</template>

<style scoped>
.face-page {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.face-page__hero {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  padding: 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface-soft);
}

.face-page__hero-title {
  font-size: 40rpx;
  font-weight: 600;
}

.face-page__hero-desc {
  font-size: 24rpx;
  line-height: 1.6;
  color: var(--gallery-muted);
}

.face-page__upload-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 80rpx 40rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 2rpx dashed var(--gallery-border);
}

.face-page__upload-icon {
  font-size: 64rpx;
  color: var(--gallery-muted);
  line-height: 1;
}

.face-page__upload-label {
  font-size: 32rpx;
  font-weight: 500;
  color: var(--gallery-text);
}

.face-page__upload-hint {
  font-size: 22rpx;
  color: var(--gallery-muted);
}

.face-page__preview-card {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.face-page__preview-image {
  width: 100%;
  height: 400rpx;
  border-radius: 16rpx;
  background: var(--gallery-surface-soft);
}

.face-page__preview-actions {
  display: flex;
  gap: 16rpx;
}

.face-page__btn-primary {
  flex: 1;
  background: var(--gallery-accent);
  color: #ffffff;
  border-radius: 999rpx;
}

.face-page__btn-secondary {
  flex: 1;
  background: var(--gallery-surface-soft);
  color: var(--gallery-text);
  border-radius: 999rpx;
}

.face-page__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx;
}

.face-page__loading-text {
  font-size: 26rpx;
  color: var(--gallery-muted);
}

.face-page__error {
  padding: 24rpx;
  border-radius: 16rpx;
  background: #fff0f0;
  color: #c0392b;
  font-size: 26rpx;
  text-align: center;
}

.face-page__result-card {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.face-page__result-title {
  font-size: 32rpx;
  font-weight: 600;
}

.face-page__result-body {
  font-size: 26rpx;
  line-height: 1.7;
  color: var(--gallery-text);
  white-space: pre-wrap;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/pages/face-reading/index.vue
git commit -m "$(cat <<'EOF'
feat(face-reading): add face-reading page

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: 注册前端页面

**Files:**
- Modify: `frontend/src/pages.json`

- [ ] **Step 1: 在 pages.json 中注册新页面**

在 `pages` 数组中添加（放在 `pages/home/index` 之后即可）：

```json
    {
      "path": "pages/face-reading/index",
      "style": {
        "navigationBarTitleText": "面相分析",
        "navigationBarBackgroundColor": "#fdf8f8",
        "backgroundColor": "#fdf8f8"
      }
    }
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/pages.json
git commit -m "$(cat <<'EOF'
feat(face-reading): register face-reading page

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: 添加首页入口

**Files:**
- Modify: `frontend/src/pages/home/index.vue`

- [ ] **Step 1: 在首页添加「面相分析」入口卡片**

在 `home-page` 的 `SceneHeroCard` 之后，第一个 `home-page__section` 之前，插入一个新的功能入口区：

```vue
      <view class="home-page__section">
        <text class="home-page__eyebrow">AI Tools</text>
        <view class="home-page__tool-card" @click="openFaceReading">
          <view class="home-page__tool-info">
            <text class="home-page__tool-title">面相分析</text>
            <text class="home-page__tool-desc">上传照片，AI 解析面相运势</text>
          </view>
          <text class="home-page__tool-arrow">›</text>
        </view>
      </view>
```

在 `<script setup>` 中新增方法：

```typescript
function openFaceReading() {
  uni.navigateTo({ url: '/pages/face-reading/index' })
}
```

在 `<style scoped>` 中新增样式：

```css
.home-page__tool-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.home-page__tool-info {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.home-page__tool-title {
  font-size: 32rpx;
  font-weight: 600;
}

.home-page__tool-desc {
  font-size: 24rpx;
  color: var(--gallery-muted);
}

.home-page__tool-arrow {
  font-size: 40rpx;
  color: var(--gallery-muted);
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/pages/home/index.vue
git commit -m "$(cat <<'EOF'
feat(home): add face-reading entry card

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 8: 集成验证

**Files:** 无新增文件，仅验证

- [ ] **Step 1: 后端编译和测试全量通过**

```bash
cd /root/project/image-play/backend
go test ./...
go build ./cmd/api
```

Expected: 全部测试通过，编译成功

- [ ] **Step 2: 前端构建通过**

```bash
cd /root/project/image-play/frontend
npm run build:mp-weixin
```

Expected: 构建成功，无 TypeScript/语法错误

- [ ] **Step 3: 最终提交（如有额外修复）**

若有修复，单独提交；若无，此步骤跳过。

---

## Self-Review

**1. Spec coverage:**
- ✅ 后端代理调用 DMXAPI → Task 2
- ✅ 配置扩展 → Task 1
- ✅ 前端页面风格一致 → Task 5（使用 GalleryPageShell 和现有 CSS 变量）
- ✅ 中文文案 → Task 5
- ✅ 入口卡片 → Task 7
- ✅ 结果复制 → Task 5
- ✅ 图片大小限制 → Task 2（后端 413）、Task 5（前端 7MB base64 检查）
- ✅ 错误处理 → Task 2、Task 5

**2. Placeholder scan:**
- 无 TBD/TODO
- 所有步骤包含完整代码和命令
- 无模糊描述

**3. Type consistency:**
- `image_base64` 前后端字段名一致
- `FaceReadingResponse` 的 `Result` 字段与前端 `res.result` 一致
- 路由路径 `/api/face-reading` 前后端一致
