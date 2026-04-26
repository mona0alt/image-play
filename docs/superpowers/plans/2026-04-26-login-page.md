# 小程序登录页实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现强制登录页流程，包含艺廊风格的微信登录页、昵称设置页，以及后端 `PUT /api/me` 接口。

**Architecture:** 启动时检查 token，未登录则跳转登录页；微信登录成功后引导设置昵称（可跳过）；运行时 token 过期通过 401 统一拦截回登录页。

**Tech Stack:** uni-app (Vue 3) 前端，Go + Gin + PostgreSQL 后端。

---

## 文件映射

| 文件 | 操作 | 说明 |
|---|---|---|
| `backend/internal/domain/user/service.go` | 修改 | Service 层添加 `GetByID` 和 `UpdateNickname` |
| `backend/internal/repository/postgres/user_repo.go` | 修改 | Repo 实现 `UpdateNickname` |
| `backend/internal/http/handlers/me_handler.go` | 新增 | `UpdateMeHandler` |
| `backend/internal/http/handlers/auth_handler_test.go` | 修改 | mock repo 添加 `UpdateNickname`，新增 handler 测试 |
| `backend/internal/http/router.go` | 修改 | 注册 `PUT /api/me` |
| `frontend/src/pages/login/index.vue` | 新增 | 登录页 |
| `frontend/src/pages/nickname-setup/index.vue` | 新增 | 昵称设置页 |
| `frontend/src/pages.json` | 修改 | 注册新页面 |
| `frontend/src/services/api.ts` | 修改 | 添加 `updateMe`，401 处理跳转登录页 |
| `frontend/src/services/session.ts` | 修改 | `ensureSession` 不再自动静默登录，无 token 时跳转登录页 |
| `frontend/src/pages/profile/index.vue` | 修改 | 修复 401 处理，不再静默重试 |
| `frontend/src/App.vue` | 修改 | 启动检查 token + 导航拦截 |

---

### Task 1: 后端 User Repo — 添加 UpdateNickname

**Files:**
- Modify: `backend/internal/domain/user/service.go`
- Modify: `backend/internal/repository/postgres/user_repo.go`
- Modify: `backend/internal/http/handlers/auth_handler_test.go`

- [ ] **Step 1: 在 Repository 接口添加 UpdateNickname 签名**

在 `backend/internal/domain/user/service.go` 的 `Repository` 接口中添加：

```go
	UpdateNickname(ctx context.Context, id int64, nickname string) error
```

- [ ] **Step 2: 在 UserRepo 实现 UpdateNickname**

在 `backend/internal/repository/postgres/user_repo.go` 末尾添加：

```go
func (r *UserRepo) UpdateNickname(ctx context.Context, id int64, nickname string) error {
	const query = `UPDATE users SET nickname = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, nickname, time.Now(), id)
	return err
}
```

- [ ] **Step 3: 在 mockUserRepo 添加 UpdateNickname**

在 `backend/internal/http/handlers/auth_handler_test.go` 的 `mockUserRepo` 中添加：

```go
func (r *mockUserRepo) UpdateNickname(_ context.Context, id int64, nickname string) error {
	if account, ok := r.usersByID[id]; ok {
		account.Nickname = nickname
	}
	return nil
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/domain/user/service.go backend/internal/repository/postgres/user_repo.go backend/internal/http/handlers/auth_handler_test.go
git commit -m "feat: add UpdateNickname to user repo and service"
```

---

### Task 2: 后端 User Service — 添加 GetByID 和 UpdateNickname

**Files:**
- Modify: `backend/internal/domain/user/service.go`

- [ ] **Step 1: 在 Service 添加方法**

在 `backend/internal/domain/user/service.go` 末尾添加：

```go
func (s *Service) GetByID(ctx context.Context, id int64) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateNickname(ctx context.Context, id int64, nickname string) error {
	return s.repo.UpdateNickname(ctx, id, nickname)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/domain/user/service.go
git commit -m "feat: add GetByID and UpdateNickname to user service"
```

---

### Task 3: 后端 UpdateMeHandler — TDD

**Files:**
- Create: `backend/internal/http/handlers/me_handler.go`
- Modify: `backend/internal/http/handlers/auth_handler_test.go`

- [ ] **Step 1: 写 UpdateMeHandler 测试**

在 `backend/internal/http/handlers/auth_handler_test.go` 末尾添加：

```go
func TestUpdateMeUpdatesNickname(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	repo.usersByID[7] = &user.User{
		ID:        7,
		OpenID:    "wx-openid-7",
		Balance:   12,
		FreeQuota: 2,
		Nickname:  "OldName",
	}
	svc := user.NewService(repo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(7))
		c.Next()
	})
	r.PUT("/me", UpdateMeHandler(svc))

	reqBody := `{"nickname":"NewName"}`
	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp User
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, int64(7), resp.ID)
	require.Equal(t, "NewName", resp.Nickname)
	require.Equal(t, int64(12), resp.Balance)
	require.Equal(t, int64(2), resp.FreeQuota)

	// Verify persisted
	updated, _ := repo.GetByID(context.Background(), 7)
	require.Equal(t, "NewName", updated.Nickname)
}

func TestUpdateMeReturnsUnauthorizedWhenMissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := user.NewService(newMockUserRepo())
	r := gin.New()
	r.PUT("/me", UpdateMeHandler(svc))

	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(`{"nickname":"Name"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMeReturnsBadRequestForInvalidNickname(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockUserRepo()
	repo.usersByID[7] = &user.User{ID: 7}
	svc := user.NewService(repo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(7))
		c.Next()
	})
	r.PUT("/me", UpdateMeHandler(svc))

	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(`{"nickname":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd backend && go test ./internal/http/handlers/ -v -run TestUpdateMe
```

Expected: FAIL with `UpdateMeHandler not defined`

- [ ] **Step 3: 实现 UpdateMeHandler**

创建 `backend/internal/http/handlers/me_handler.go`：

```go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"image-play/internal/domain/user"
)

func UpdateMeHandler(userSvc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("user_id")
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			Nickname string `json:"nickname" binding:"required,min=2,max=20"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := userSvc.UpdateNickname(c.Request.Context(), uid, req.Nickname); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}

		account, err := userSvc.GetByID(c.Request.Context(), uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
			return
		}
		if account == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, User{
			ID:        account.ID,
			Nickname:  account.Nickname,
			Balance:   int64(account.Balance),
			FreeQuota: int64(account.FreeQuota),
		})
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd backend && go test ./internal/http/handlers/ -v -run TestUpdateMe
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/http/handlers/me_handler.go backend/internal/http/handlers/auth_handler_test.go
git commit -m "feat: add UpdateMe handler with tests"
```

---

### Task 4: 后端 Router — 注册 PUT /api/me

**Files:**
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: 注册路由**

在 `backend/internal/http/router.go` 的 `authorized.GET("/me", ...)` 下方添加：

```go
	authorized.PUT("/me", handlers.UpdateMeHandler(userSvc))
```

- [ ] **Step 2: 编译检查**

```bash
cd backend && go build ./...
```

Expected: 编译成功，无错误。

- [ ] **Step 3: Commit**

```bash
git add backend/internal/http/router.go
git commit -m "feat: wire PUT /api/me route"
```

---

### Task 5: 前端登录页

**Files:**
- Create: `frontend/src/pages/login/index.vue`
- Modify: `frontend/src/pages.json`

- [ ] **Step 1: 创建登录页**

创建 `frontend/src/pages/login/index.vue`：

```vue
<script setup lang="ts">
import { ref } from 'vue'
import { login } from '../../services/api'

const agreed = ref(false)
const loading = ref(false)

async function handleWechatLogin() {
  if (!agreed.value) {
    uni.vibrateShort()
    uni.showToast({ title: '请先同意用户协议', icon: 'none' })
    return
  }

  loading.value = true
  try {
    const loginRes = await uni.login({ provider: 'weixin' })
    if (!loginRes.code) {
      throw new Error('WeChat login failed: no code')
    }
    const res = await login(loginRes.code)
    uni.setStorageSync('access_token', res.access_token)
    uni.reLaunch({ url: '/pages/nickname-setup/index' })
  } catch (err) {
    console.error('[login] failed:', err)
    uni.showToast({ title: '微信登录失败，请重试', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function showToast(msg: string) {
  uni.showToast({ title: msg, icon: 'none' })
}
</script>

<template>
  <view class="login-page">
    <view class="brand-section">
      <view class="logo">
        <view class="logo-square"></view>
        <view class="logo-inner"></view>
        <text class="logo-icon">&#x25A1;</text>
      </view>
      <text class="brand-title">精品场景馆</text>
      <text class="brand-subtitle">开启您的艺术创作之旅</text>
    </view>

    <view class="actions">
      <button class="btn-wechat" :disabled="loading" @click="handleWechatLogin">
        <text class="wechat-icon">&#x263A;</text>
        <text>{{ loading ? '登录中...' : '微信一键登录' }}</text>
      </button>

      <view class="gallery-preview">
        <view class="preview-item">
          <view class="preview-placeholder"></view>
        </view>
        <view class="preview-item shifted">
          <view class="preview-placeholder"></view>
        </view>
        <view class="preview-item">
          <view class="preview-placeholder"></view>
        </view>
      </view>
    </view>

    <view class="footer">
      <view class="checkbox-row" @click="agreed = !agreed">
        <view class="custom-checkbox" :class="{ checked: agreed }"></view>
        <text class="terms-text">
          我已阅读并同意
          <text class="link" @click.stop="showToast('协议内容即将上线')">《用户协议》</text>
          和
          <text class="link" @click.stop="showToast('协议内容即将上线')">《隐私政策》</text>
        </text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 0 64rpx 48rpx;
  background: var(--gallery-bg);
}

.brand-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24rpx;
}

.logo {
  width: 192rpx;
  height: 192rpx;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-square {
  position: absolute;
  inset: 0;
  border: 1rpx solid #1c1b1b;
  transform: rotate(45deg);
}

.logo-inner {
  position: absolute;
  inset: 32rpx;
  border: 1rpx solid rgba(28, 27, 27, 0.2);
  transform: rotate(-12deg);
}

.logo-icon {
  font-size: 64rpx;
  color: #1c1b1b;
  z-index: 1;
}

.brand-title {
  font-size: 64rpx;
  font-weight: 600;
  color: #1c1b1b;
  letter-spacing: 0.03em;
}

.brand-subtitle {
  font-size: 24rpx;
  color: #6d6865;
  letter-spacing: 0.1em;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 48rpx;
  margin-bottom: 48rpx;
}

.btn-wechat {
  height: 112rpx;
  background: #000000;
  color: #ffffff;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  font-size: 32rpx;
  font-weight: 600;
}

.btn-wechat:active {
  transform: scale(0.98);
}

.wechat-icon {
  font-size: 40rpx;
}

.gallery-preview {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16rpx;
  opacity: 0.4;
}

.preview-item {
  padding: 8rpx;
  border: 1rpx solid #e5e2e1;
}

.preview-item.shifted {
  margin-top: 32rpx;
}

.preview-placeholder {
  height: 200rpx;
  background: #e5e2e1;
}

.footer {
  padding-top: 24rpx;
}

.checkbox-row {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
}

.custom-checkbox {
  width: 32rpx;
  height: 32rpx;
  border: 2rpx solid #1c1b1b;
  flex-shrink: 0;
  margin-top: 4rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.custom-checkbox.checked::after {
  content: '';
  width: 16rpx;
  height: 16rpx;
  background: #1c1b1b;
}

.terms-text {
  font-size: 22rpx;
  color: #6d6865;
  line-height: 1.6;
}

.link {
  color: #1c1b1b;
  text-decoration: underline;
  text-underline-offset: 8rpx;
}
</style>
```

- [ ] **Step 2: 注册页面**

在 `frontend/src/pages.json` 的 `pages` 数组中添加（不放在第一个，避免已登录用户闪屏）：

```json
    {
      "path": "pages/login/index",
      "style": {
        "navigationBarTitleText": "SCENE GALLERY",
        "navigationBarBackgroundColor": "#fdf8f8",
        "backgroundColor": "#fdf8f8"
      }
    },
    {
      "path": "pages/nickname-setup/index",
      "style": {
        "navigationBarTitleText": "设置昵称",
        "navigationBarBackgroundColor": "#fdf8f8",
        "backgroundColor": "#fdf8f8"
      }
    },
```


- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/login/index.vue frontend/src/pages.json
git commit -m "feat: add login page with gallery style"
```

---

### Task 6: 前端昵称设置页

**Files:**
- Create: `frontend/src/pages/nickname-setup/index.vue`

- [ ] **Step 1: 创建昵称设置页**

创建 `frontend/src/pages/nickname-setup/index.vue`：

```vue
<script setup lang="ts">
import { ref } from 'vue'
import { updateMe } from '../../services/api'

const nickname = ref('')
const loading = ref(false)

async function handleConfirm() {
  const trimmed = nickname.value.trim()
  if (trimmed.length < 2 || trimmed.length > 20) {
    uni.showToast({ title: '昵称长度需在 2-20 个字符之间', icon: 'none' })
    return
  }

  loading.value = true
  try {
    await updateMe(trimmed)
    uni.reLaunch({ url: '/pages/home/index' })
  } catch (err) {
    console.error('[nickname] update failed:', err)
    uni.showToast({ title: '更新失败，请重试', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function handleSkip() {
  uni.reLaunch({ url: '/pages/home/index' })
}

function goBack() {
  uni.navigateBack()
}
</script>

<template>
  <view class="nickname-page">
    <view class="header">
      <text class="back-arrow" @click="goBack">&#x2190;</text>
    </view>

    <view class="content">
      <text class="title">如何称呼您？</text>
      <text class="subtitle">您可以随时在个人中心修改</text>

      <view class="input-wrapper">
        <input
          v-model="nickname"
          class="nickname-input"
          type="text"
          placeholder="请输入昵称"
          maxlength="20"
        />
        <text class="char-count">{{ nickname.length }}/20</text>
      </view>
    </view>

    <view class="actions">
      <button class="btn-confirm" :disabled="loading" @click="handleConfirm">
        {{ loading ? '保存中...' : '确认' }}
      </button>
      <text class="btn-skip" @click="handleSkip">跳过</text>
    </view>
  </view>
</template>

<style scoped>
.nickname-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 0 64rpx 48rpx;
  background: var(--gallery-bg);
}

.header {
  padding-top: 24rpx;
  height: 88rpx;
  display: flex;
  align-items: center;
}

.back-arrow {
  font-size: 40rpx;
  color: #1c1b1b;
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  margin-top: 48rpx;
}

.title {
  font-size: 48rpx;
  font-weight: 500;
  color: #1c1b1b;
}

.subtitle {
  font-size: 28rpx;
  color: #6d6865;
}

.input-wrapper {
  margin-top: 64rpx;
  position: relative;
}

.nickname-input {
  width: 100%;
  height: 96rpx;
  font-size: 36rpx;
  color: #1c1b1b;
  border-bottom: 1rpx solid #c4c7c7;
  background: transparent;
}

.nickname-input:focus {
  border-bottom-color: #1c1b1b;
}

.char-count {
  position: absolute;
  right: 0;
  bottom: -40rpx;
  font-size: 24rpx;
  color: #6d6865;
}

.actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 32rpx;
  margin-top: 48rpx;
}

.btn-confirm {
  width: 100%;
  height: 96rpx;
  background: #000000;
  color: #ffffff;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
}

.btn-confirm:active {
  transform: scale(0.98);
}

.btn-skip {
  font-size: 28rpx;
  color: #1c1b1b;
  text-decoration: underline;
  text-underline-offset: 8rpx;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/nickname-setup/index.vue
git commit -m "feat: add nickname setup page"
```

---

### Task 7: 前端 API 服务更新

**Files:**
- Modify: `frontend/src/services/api.ts`
- Modify: `frontend/src/services/session.ts`
- Modify: `frontend/src/pages/profile/index.vue`

- [ ] **Step 1: 添加 updateMe 函数**

在 `frontend/src/services/api.ts` 中，在 `getMe()` 下方添加：

```typescript
export function updateMe(nickname: string) {
  return request<{ id: number; nickname: string; balance: number; free_quota: number }>({
    url: '/api/me',
    method: 'PUT',
    data: { nickname },
    headers: { 'Content-Type': 'application/json' },
  })
}
```

- [ ] **Step 2: 修改 request 的 401 处理**

在 `frontend/src/services/api.ts` 的 `request` 函数中，将 401 处理改为：

```typescript
        if (response.statusCode === 401) {
          uni.removeStorageSync('access_token')
          uni.reLaunch({ url: '/pages/login/index' })
          reject(new Error('Unauthorized'))
          return
        }
```

- [ ] **Step 3: 修改 ensureSession 不再自动静默登录**

在 `frontend/src/services/session.ts` 中，将 `ensureSession()` 替换为：

```typescript
export function ensureSession(): Promise<string> {
  const existing = uni.getStorageSync('access_token') as string | undefined
  if (existing) {
    return Promise.resolve(existing)
  }
  uni.reLaunch({ url: '/pages/login/index' })
  return Promise.reject(new Error('Unauthorized'))
}
```

- [ ] **Step 4: 修复 profile 页面的 401 处理**

在 `frontend/src/pages/profile/index.vue` 的 `loadProfilePage` catch 块中，将：

```typescript
    if (e.message === 'Unauthorized') {
      userStore.clear()
      await ensureSession()
      return loadProfilePage()
    }
```

改为：

```typescript
    if (e.message === 'Unauthorized') {
      userStore.clear()
      // request() 已处理跳转，不再静默重试
      return
    }
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/services/api.ts frontend/src/services/session.ts frontend/src/pages/profile/index.vue
git commit -m "feat: add updateMe API, 401 redirect, and fix session flow"
```

---

### Task 8: 前端 App.vue 启动逻辑 + 导航拦截

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: 修改 App.vue**

将 `frontend/src/App.vue` 替换为：

```vue
<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'

onLaunch(() => {
  const token = uni.getStorageSync('access_token') as string | undefined
  if (!token) {
    uni.reLaunch({ url: '/pages/login/index' })
  }

  // 导航拦截：未登录时阻止跳转到非登录页
  const methods = ['navigateTo', 'redirectTo', 'switchTab', 'reLaunch'] as const
  methods.forEach((method) => {
    uni.addInterceptor(method, {
      invoke(args) {
        const token = uni.getStorageSync('access_token') as string | undefined
        const url = (args as any).url || ''
        if (!token && !url.includes('/pages/login')) {
          uni.reLaunch({ url: '/pages/login/index' })
          return false
        }
      },
    })
  })
})
</script>

<style>
page {
  --gallery-bg: #fdf8f8;
  --gallery-surface: #ffffff;
  --gallery-surface-soft: #f4efed;
  --gallery-border: rgba(28, 27, 27, 0.08);
  --gallery-text: #1c1b1b;
  --gallery-muted: #6d6865;
  --gallery-accent: #111111;
  background: var(--gallery-bg);
  color: var(--gallery-text);
}

view,
text,
button,
image,
scroll-view,
input,
textarea {
  box-sizing: border-box;
}

button {
  border: none;
}

button::after {
  border: none;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/App.vue
git commit -m "feat: enforce login on launch and add navigation guard"
```

---

## 验证清单

- [ ] 后端测试全部通过：`cd backend && go test ./...`
- [ ] 后端编译通过：`cd backend && go build ./...`
- [ ] 前端编译通过：`cd frontend && npm run dev:mp-weixin`（无报错）
- [ ] 登录流程走通：清除 token → 启动 → 看到登录页 → 微信登录 → 昵称设置 → 进入首页
- [ ] 401 处理：登录后手动删除 token → 触发 API 请求 → 自动回到登录页
- [ ] 导航拦截：未登录时尝试跳转其他页面 → 被拦截回登录页

## Spec 覆盖检查

| Spec 要求 | 对应 Task |
|---|---|
| 启动后强制显示登录页 | Task 8 |
| 登录页品牌区 + 微信一键登录 + 协议勾选 | Task 5 |
| 微信登录成功后跳转昵称设置页 | Task 5 |
| 昵称设置页（输入/跳过） | Task 6 |
| 后端 `PUT /api/me` 接口 | Task 1-4 |
| 401 统一处理回登录页 | Task 7 |
| 导航拦截 | Task 8 |
| 艺廊风格样式 | Task 5, 6 |
| ensureSession 不再自动静默登录 | Task 7 |
| 修复 profile 页面 401 循环 | Task 7 |
