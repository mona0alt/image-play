# image-play

AI 图片场景生成小程序。uni-app 微信小程序前端 + Go 后端。

## 项目结构

- `backend/` — Go 后端（Gin + PostgreSQL），包含 API 服务和异步 Worker
- `frontend/` — uni-app 微信小程序前端（Vue 3）
- `infra/` — Docker Compose 配置和 SQL 种子文件

## 技术栈

- **后端**：Go, Gin, PostgreSQL, JWT, 微信登录
- **前端**：Vue 3, uni-app, TypeScript
- **构建**：Make, Go modules, npm

### 大模型交互技术选型

- **LLM 应用框架**: [CloudWeGo Eino](https://github.com/cloudwego/eino) — 构建大语言模型应用的 Go 框架，提供统一的组件抽象、流式处理、回调机制。
- **文本模型组件**: `eino-ext/components/model/openai` — 兼容 OpenAI API 格式的 ChatModel 实现。Moonshot 等国内主流模型均兼容此格式。
- **图片生成**: 原生 HTTP 封装 — eino 生态暂无标准文生图组件，直接封装更可控。
- **模块位置**: `backend/internal/infrastructure/llm/`
- **接口设计**: 项目自定义 `TextClient` / `ImageClient` 接口，完全封装 eino 细节，调用方零依赖。

### 前端 SSE 流式接收（微信小程序）

面相分析等长文本生成场景使用 SSE（Server-Sent Events）流式输出，避免用户长时间空白等待。微信小程序不支持浏览器标准的 `EventSource`，需使用原生 `wx.request` 的 chunked 传输能力实现。

**实现要点：**

- **必须使用 `wx.request` 而非 `uni.request`**：uni-app Vue3 中 `uni.request` 返回 Promise，无法获取 `RequestTask` 对象，因此无法注册 `onChunkReceived` 监听。直接使用微信小程序原生 `wx.request` 才能拿到 `RequestTask`。
- **`enableChunked: true`**：开启 HTTP 分块传输，配合后端 `c.Writer.Flush()` 实现实时推送。
- **`responseType: 'arraybuffer'`**：必须显式设置，否则 `onChunkReceived` 回调中的 `res.data` 类型不确定。
- **UTF-8 解码兼容**：微信小程序不支持 `TextDecoder`，需使用 `Uint8Array` + `String.fromCharCode` + `decodeURIComponent(escape())` 解码 ArrayBuffer。
- **SSE 协议解析**：`onChunkReceived` 返回的是原始 HTTP chunk，不是完整 SSE 消息。需维护 buffer，按 `\n\n` 分隔后提取 `data: ` 前缀的内容，再 JSON 解析。
- **`uni.canIUse` 不可靠**：`uni.canIUse("request.enableChunked")` 在 Windows 版微信开发者工具模拟器上可能误报为 `false`，应以实际 `requestTask.onChunkReceived` 是否存在为准。
- **Fallback 兜底**：若流式不可用，自动回退到普通 HTTP 请求，timeout 设为 60 秒（避免 LLM 生成超过 15 秒导致超时失败）。
- **相关文件**：`frontend/src/services/api.ts`（`faceReadingStream` 函数 + `SSEParser` 类）、`backend/internal/http/handlers/face_reading_handler.go`（SSE 响应输出）。

## 前置依赖

- Go 1.22+
- Node.js 18+
- PostgreSQL（本地或 Docker）
- 微信开发者工具（前端运行需要）

## 快速开始

### 1. 启动数据库

```bash
make dev-db
```

或使用 Docker：

```bash
cd infra && docker compose up -d postgres
```

### 2. 启动后端

项目根目录提供了 `start-backend.sh` 脚本，会自动杀掉端口占用进程并启动服务：

```bash
# 启动 API 服务（默认）
./start-backend.sh api

# 启动 Worker
./start-backend.sh worker

# 同时启动 API + Worker
./start-backend.sh all
```

也可使用 Make 命令：

```bash
make dev-api     # 启动 API 服务
make dev-worker  # 启动 Worker
```

API 默认监听 `:8080`，配置在 `backend/config.yaml`，支持环境变量覆盖。

### 3. 构建前端

```bash
make dev-frontend
```

构建产物在 `frontend/dist/build/mp-weixin/`。在微信开发者工具中导入该目录即可运行。

## 常用 Make 命令

| 命令 | 说明 |
|---|---|
| `make build` | 构建全部（后端二进制 + 前端） |
| `make build-backend` | 仅构建后端，输出到 `dist/api` 和 `dist/worker` |
| `make build-frontend` | 仅构建前端（含 `npm install`） |
| `make dev-api` | 启动 API 开发服务器 |
| `make dev-worker` | 启动 Worker 开发服务器 |
| `make dev-db` | 启动 PostgreSQL |
| `make dev-frontend` | 前端开发构建 |
| `make test` | 运行全部测试 |
| `make clean` | 清理构建产物和依赖 |

## 配置说明

### 后端配置

`backend/config.yaml`：

```yaml
app_env: development
port: "8080"
database_url: "postgres://postgres:postgres@localhost:5432/image_play?sslmode=disable"
jwt_secret: "dev-secret-change-me-in-production"
wechat_app_id: "your-app-id"
wechat_app_secret: "your-app-secret"
```

支持环境变量覆盖：`APP_ENV`, `PORT`, `DATABASE_URL`, `JWT_SECRET`, `WECHAT_APP_ID`, `WECHAT_APP_SECRET`。

### 前端配置

前端后端地址在 `frontend/.env` 中配置：

```bash
VITE_API_BASE=http://192.168.0.105:8080
```

开发时改为本机局域网 IP，以便微信开发者工具/真机访问。

## 微信小程序开发注意事项

- 在微信开发者工具中勾选「不校验合法域名、web-view...」以支持 IP 地址调试
- 真机调试需确保手机和电脑在同一局域网
- 生产环境需在小程序后台配置 HTTPS 服务器域名
