# 微信登录接入设计文档

## 1. 背景与目标

将现有 mock 登录彻底替换为真实微信小程序登录，获取 openid，与当前用户体系打通。

- 删除所有 mock 登录代码
- 后端配置文件管理 `appid` + `appsecret`
- 新用户自动初始化（nickname = "创作者" + 随机3位数字，头像默认）
- 仅获取 openid，不获取用户昵称/头像授权

## 2. 整体数据流

```
小程序启动
    │
    ▼
wx.login() ──code──▶ 前端
    │
    ▼
POST /api/auth/login { code }
    │
    ▼
后端 ──code+appid+secret──▶ 微信 jscode2session
    │
    ◀──openid + session_key──
    │
    ▼
查 users 表：openid 存在？────Y──▶ 返回已有用户
    │                              生成 JWT
    N
    ▼
新建用户：nickname="创作者"+随机3位
    │
    ▼
返回 { access_token, user }
    │
    ▼
前端存 token，后续请求带 Authorization: Bearer <token>
```

## 3. 后端设计

### 3.1 配置变更 (`internal/config/config.go`)

新增两个字段：

```go
type Config struct {
    AppEnv          string
    Port            string
    DatabaseURL     string
    JWTSecret       string
    WechatAppID     string
    WechatAppSecret string
}
```

配置文件（或环境变量）中需提供 `wechat.appid` 和 `wechat.appsecret`。

### 3.2 微信客户端 (`internal/infrastructure/wechat/client.go`)

职责：封装对微信 `jscode2session` 的 HTTP 调用。

```go
type Code2SessionResponse struct {
    OpenID     string `json:"openid"`
    SessionKey string `json:"session_key"`
    UnionID    string `json:"unionid"`
    ErrCode    int    `json:"errcode"`
    ErrMsg     string `json:"errmsg"`
}

func (c *Client) Code2Session(ctx context.Context, code string) (*Code2SessionResponse, error)
```

实现细节：
- 请求 URL: `https://api.weixin.qq.com/sns/jscode2session`
- 参数: `appid`, `secret`, `js_code`, `grant_type=authorization_code`
- 超时: 5 秒
- 如果 `errcode != 0`，返回错误（前端应引导重新登录）

### 3.3 用户服务变更 (`internal/domain/user/service.go`)

**删除** `GetOrCreateByMockCode`。

**新增** `GetOrCreateByWxCode`：

```go
func (s *Service) GetOrCreateByWxCode(ctx context.Context, code string, wxClient *wechat.Client) (*User, bool, error)
```

流程：
1. 调用 `wxClient.Code2Session(code)` 获取 `openid`
2. `s.repo.GetByOpenID(ctx, openid)` 查库
3. 若存在，返回 `(user, false, nil)`
4. 若不存在，创建新用户：
   - `OpenID = openid`
   - `Nickname = "创作者" + rand.Intn(900)+100`（100~999）
   - `AvatarURL = ""`（默认头像由前端决定）
   - `Balance = 0`, `FreeQuota = 3`
   - 返回 `(user, true, nil)`
5. 若创建时发生唯一键冲突，再次 `GetByOpenID` 并返回（防御并发）

### 3.4 登录接口 (`internal/http/handlers/auth_handler.go`)

`LoginHandler` 保持入参出参结构不变，内部逻辑改为：

1. 绑定 `LoginRequest { Code string }`
2. 调用 `userSvc.GetOrCreateByWxCode(ctx, req.Code, wxClient)`
3. 生成 JWT：`sub = user.ID`, `exp = now + 30 days`
4. 返回 `LoginResponse`：
   - `access_token`
   - `user: { id, balance, free_quota }`（**不含 openid**）

### 3.5 JWT 有效期调整

- 由 7 天改为 **30 天**。

## 4. 前端设计

### 4.1 `services/session.ts` 重写

```ts
export async function ensureSession(): Promise<string> {
  const existing = uni.getStorageSync('access_token') as string | undefined
  if (existing) return existing

  const [err, res] = await uni.login({ provider: 'weixin' })
  if (err || !res.code) {
    throw new Error('微信登录失败')
  }

  const loginRes = await login(res.code)
  uni.setStorageSync('access_token', loginRes.access_token)
  return loginRes.access_token
}
```

### 4.2 `App.vue` 调用改名

```ts
import { ensureSession } from './services/session'
onLaunch(() => { void ensureSession() })
```

### 4.3 登录失败兜底

- `ensureSession` 失败**不阻断**小程序启动，首页正常展示
- 用户触发需要鉴权的操作（如生成图片）时，若检测到无 token 或收到 401，重新调登录

## 5. 错误处理

| 场景 | 后端 HTTP 码 | 错误码 | 前端行为 |
|------|-------------|--------|---------|
| 微信返回 code 无效/过期 | 400 | `WECHAT_LOGIN_FAILED` | 提示"登录失败，请重试"，重新 `wx.login` |
| 微信接口超时/网络异常 | 503 | `WECHAT_UNAVAILABLE` | 提示"网络不稳定"，可自动重试1次 |
| 后端数据库异常 | 500 | `SYSTEM_ERROR` | 提示"服务异常" |
| 前端未传 code | 400 | `INVALID_PARAM` | 不传（调用处保证） |

## 6. 安全考量

- `session_key` 仅在后端与微信之间流转，**不返回给前端，不落库**
- `openid` **不回显给前端**，`LoginResponse.User` 中移除 `openid` 字段
- `code` 为一次性凭证，5 分钟过期，后端不做缓存
- JWT 泄露风险由 30 天有效期控制，未来可考虑增加刷新机制

## 7. 改动文件清单

### 后端
1. `internal/config/config.go` — 新增微信配置字段
2. `internal/infrastructure/wechat/client.go` — 新增（微信 HTTP 客户端）
3. `internal/domain/user/service.go` — 删除 mock 方法，新增微信登录方法
4. `internal/http/handlers/auth_handler.go` — 替换为真实登录逻辑，调整 JWT 有效期
5. `cmd/api/main.go` — 注入微信 client 和读取配置

### 前端
1. `frontend/src/services/session.ts` — 重写为真实微信登录
2. `frontend/src/App.vue` — 调用名改为 `ensureSession`
3. `frontend/src/services/api.ts` — 移除 `openid` 从 login 返回类型（如有）

## 8. 非目标

- 不获取用户手机号（需要 `session_key` 解密，MVP 阶段不做）
- 不做 unionid 关联（单小程序场景暂时不需要）
- 不引入 JWT 刷新机制（30 天有效期足够 MVP 使用）
