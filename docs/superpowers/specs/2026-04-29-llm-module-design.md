# LLM 通用模块设计文档

## 背景

后端目前有两处与大模型的交互：

1. **面相分析** (`face_reading_handler.go`)：手写 HTTP 调用 Moonshot API，流式 SSE 返回，支持图文多模态输入。
2. **图片生成** (`generation_job.go`)：通过 `ModelClient` 接口调用，当前为 Mock，需接入真实文生图 Provider。

技术选型为 CloudWeGo **Eino** 框架（`github.com/cloudwego/eino`），文本模型使用 `eino-ext/components/model/openai`（Moonshot 兼容 OpenAI 格式）。图片生成由于 eino 暂无对应组件，使用原生 HTTP 封装。

---

## 目标

- 将所有后端大模型交互封装到单一通用模块。
- 对外完全隐藏 eino 实现细节，调用方不感知框架存在。
- 支持文本/多模态对话（含流式）和图片生成两种能力。
- 配置化驱动，单一 Provider 即可，但预留未来扩展空间。

---

## 非目标

- 不实现多 Provider 动态切换（如 OpenAI / Claude / Moonshot 同时配置并运行时选择）。
- 不实现通用的 Tool Calling 抽象（当前无此场景）。
- 不替换现有的内容审核（AuditClient）接口。

---

## 架构

```
┌─────────────────────────────────────┐
│  face_reading_handler.go            │
│  generation_job.go                  │
└──────────┬──────────────────────────┘
           │ llm.TextClient / llm.ImageClient
           ▼
┌─────────────────────────────────────┐
│  internal/infrastructure/llm        │
│  ├─ interfaces.go      (公共 API)   │
│  ├─ text_client.go     (eino 实现)  │
│  ├─ image_client.go    (HTTP 实现)  │
│  └─ convert.go         (消息转换)   │
└─────────────────────────────────────┘
```

---

## 公共 API

```go
package llm

type PartType string

const (
    PartTypeText  PartType = "text"
    PartTypeImage PartType = "image"
)

type Part struct {
    Type    PartType
    Content string // 文本内容，或图片 URL / data URL
}

type Message struct {
    Role  string // user, assistant, system
    Parts []Part
}

// StreamReader 流式读取模型输出
// Recv 返回 io.EOF 表示流正常结束
// 调用方必须在使用完成后调用 Close
 type StreamReader interface {
    Recv() (Chunk, error)
    Close()
}

type Chunk struct {
    Content string
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
```

构造函数：

```go
func NewTextClient(cfg TextConfig) (TextClient, error)
func NewImageClient(cfg ImageConfig) (ImageClient, error)
```

配置结构：

```go
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

---

## 内部实现

### textClient

内部持有 `model.ChatModel`（通过 `eino-ext/components/model/openai` 创建）。

- `Chat()`：调用 `chatModel.Generate()`，从返回的 `*schema.Message` 中提取 `Content` 字段。
- `ChatStream()`：调用 `chatModel.Stream()`，将 `schema.StreamReader[*schema.Message]` 包装为 `llm.StreamReader`。内部每次 `Recv()` 读取到 `*schema.Message` 后，提取 `Content` 组装为 `llm.Chunk`。
- 消息转换：
  - `llm.Message` 的 `Parts` 转换为 `schema.ChatMessagePart` 数组。
  - `PartTypeText` → `schema.ChatMessagePartTypeText`
  - `PartTypeImage` → `schema.ChatMessagePartTypeImageURL`，`Content` 作为 `ImageURL.URL`
  - 转换后的 `MultiContent` 赋值给 `schema.Message`，`Role` 映射为 `schema.RoleType`。

### imageClient

使用 `net/http` 直接调用文生图 API。

- 根据配置构造 HTTP POST 请求，JSON body 包含 `prompt`、`model` 等字段。
- 解析响应 JSON，提取图片 URL 返回。
- 超时、重试通过配置的 `Timeout` 和自定义 `http.Client` 控制。

---

## 配置变更

`config.yaml` 新增 `llm` 区块：

```yaml
llm:
  text:
    api_key: "sk-..."
    base_url: "https://api.moonshot.cn/v1"
    model: "kimi-k2.6"
    timeout: 300
  image:
    api_key: "..."
    base_url: "..."
    model: "..."
    timeout: 60
```

`config.go` 新增 `LLMConfig` 结构，并在 `Load()` 中实现向下兼容：若 `llm.text.api_key` 为空，则回退到旧的 `dmx_api_key` / `dmx_api_base_url` / `dmx_model`。

---

## 迁移计划

| 文件 | 变更 |
|---|---|
| `go.mod` | 引入 `github.com/cloudwego/eino` 和 `github.com/cloudwego/eino-ext/components/model/openai` |
| `internal/infrastructure/llm/` | 新建包，含接口、实现、消息转换、单元测试 |
| `internal/config/config.go` | 新增 `LLMConfig`、`TextConfig`、`ImageConfig`；保留 DMX fallback |
| `internal/http/router.go` | 接收 `llm.TextClient` 和 `llm.ImageClient` 参数并注入 Handler/Job |
| `internal/http/handlers/face_reading_handler.go` | 注入 `llm.TextClient`；构建 `[]llm.Message`；调用 `ChatStream()` 并迭代写入 SSE |
| `internal/worker/jobs/generation_job.go` | 将 `ModelClient` 接口替换为 `llm.ImageClient`（签名一致） |
| `cmd/api/main.go` | `llm.NewTextClient(cfg.LLM.Text)` → 传入 Router |
| `cmd/worker/main.go` | `llm.NewImageClient(cfg.LLM.Image)` → 传入 `GenerationJob` |

---

## 错误处理

- 所有错误使用 `fmt.Errorf("llm: ...: %w", err)` 包装。
- `ChatStream` 中流中断时通过 `Recv()` 返回错误，调用方（Handler）负责终止 SSE 连接。
- 超时按客户端级别配置，不设置全局默认值。

---

## 测试策略

- `llm` 包内对 `toSchemaMessages()` 做单元测试，验证多模态消息转换正确。
- `textClient` 和 `imageClient` 通过 `httptest.Server` mock HTTP 响应进行测试。
- Handler 和 Job 的现有测试改为 mock `llm.TextClient` / `llm.ImageClient`。

---

## 技术选型说明

| 层级 | 选型 | 理由 |
|---|---|---|
| LLM 应用框架 | Eino (CloudWeGo) | 项目已引用 eino/eino-ext 作为基础依赖，生态成熟，支持流式、多模态、回调等 |
| 文本模型组件 | eino-ext/components/model/openai | Moonshot / 大多数国内模型均兼容 OpenAI API 格式，可直接复用 |
| 图片生成 | 原生 HTTP 封装 | eino 生态暂无文生图标准组件，直接封装更可控 |
| 接口风格 | 项目自定义接口 | 完全封装 eino，未来替换框架时调用方零改动 |
