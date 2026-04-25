# AI 图片场景馆边界修复 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把当前项目从 mock 与真实链路并存的状态修正为“登录、模板、生成、历史/结果”主流程真实可用。

**Architecture:** 保持现有 `UniApp + Vue3 + Go API + Go Worker + PostgreSQL` 结构不变，优先收紧 `auth/user`、`template/config`、`generation/history` 三条边界。后端建立真实用户落库与模板查询能力，生成链在创建和 worker 执行阶段都只消费真实模板数据；前端取消本地 mock 模板和伪提交流程，改为通过真实接口驱动页面和状态。

**Tech Stack:** UniApp、Vue 3、TypeScript、Pinia、Vitest、Go 1.22、Gin、PostgreSQL、database/sql

---

## File Structure

### Backend

- Create: `backend/internal/domain/user/service.go`
  负责 mock 登录场景下的用户查找/创建边界。
- Create: `backend/internal/domain/user/service_test.go`
  覆盖首次登录创建用户与重复登录复用用户。
- Create: `backend/internal/repository/postgres/user_repo.go`
  提供 users 表的读取与创建。
- Create: `backend/internal/repository/postgres/scene_template_repo.go`
  提供 `scene_templates` 的列表与单模板查询。
- Create: `backend/internal/http/handlers/templates_handler.go`
  暴露模板查询接口。
- Create: `backend/internal/http/handlers/templates_handler_test.go`
  覆盖场景模板查询行为。
- Create: `backend/internal/worker/jobs/generation_job_test.go`
  覆盖 worker 使用统一 Prompt Builder 和状态推进。
- Modify: `backend/internal/http/router.go`
  注入 user/template 依赖并注册新接口。
- Modify: `backend/internal/http/handlers/auth_handler.go`
  登录改为真实用户查找/创建，`/me` 改用用户边界。
- Modify: `backend/internal/http/handlers/auth_handler_test.go`
  覆盖登录落库行为。
- Modify: `backend/internal/http/handlers/config_handler.go`
  客户端配置返回真实五个场景顺序。
- Modify: `backend/internal/http/handlers/generations_handler.go`
  保持接口不变，但配合模板校验返回明确错误。
- Modify: `backend/internal/http/handlers/generations_handler_test.go`
  覆盖禁用模板/合法模板创建。
- Modify: `backend/internal/http/handlers/admin_metrics_handler.go`
  成功统计口径统一为 `success`。
- Modify: `backend/internal/http/handlers/admin_metrics_handler_test.go`
  更新成功状态口径测试。
- Modify: `backend/internal/domain/scenes/catalog.go`
  定义合法场景集合和统一顺序。
- Modify: `backend/internal/domain/scenes/templates.go`
  改为模板实体与 prompt preset 结构定义，不再持有静态模板真源。
- Modify: `backend/internal/domain/scenes/prompt_builder.go`
  改为基于模板 preset 构建 prompt。
- Modify: `backend/internal/domain/scenes/prompt_builder_test.go`
  覆盖基于 preset 的 prompt 构建。
- Modify: `backend/internal/domain/generation/service.go`
  创建任务前校验模板是否合法且启用。
- Modify: `backend/internal/domain/generation/service_test.go`
  覆盖模板校验和合法任务创建。
- Modify: `backend/internal/domain/generation/memory_repo_test.go`
  为模板校验增加测试桩。
- Modify: `backend/internal/worker/jobs/generation_job.go`
  通过模板 repo + Prompt Builder 生成统一 prompt。
- Modify: `infra/sql/seed_scene_templates.sql`
  补齐五个场景的最小模板与 schema/prompt preset。

### Frontend

- Create: `frontend/src/pages.json`
  定义 UniApp 页面入口。
- Create: `frontend/src/services/session.ts`
  提供 mock 会话初始化，避免依赖不存在的登录页。
- Create: `frontend/src/services/__tests__/api.test.ts`
  覆盖模板和历史响应映射。
- Create: `frontend/src/utils/generation.ts`
  提供状态判断、历史筛选、按 ID 查找结果。
- Create: `frontend/src/utils/__tests__/generation.test.ts`
  覆盖状态与历史纯逻辑。
- Modify: `frontend/src/main.ts`
  安装 Pinia。
- Modify: `frontend/src/App.vue`
  启动时初始化 mock 会话。
- Modify: `frontend/src/services/api.ts`
  新增模板、生成接口和响应映射，401 回首页。
- Modify: `frontend/src/store/generation.ts`
  管理模板加载、提交状态与最近一次 `generation_id`。
- Modify: `frontend/src/store/__tests__/generation.test.ts`
  覆盖加载模板与真实提交流程。
- Modify: `frontend/src/pages/home/index.vue`
  修正内容/错误/加载态分支。
- Modify: `frontend/src/pages/scene/index.vue`
  移除 `mockTemplates`，改用真实模板和真实提交。
- Modify: `frontend/src/pages/result/index.vue`
  只依赖 `generation_id` + 历史数据恢复结果，并在活动状态下轮询历史。
- Modify: `frontend/src/pages/history/index.vue`
  使用统一状态筛选与跳转参数。
- Modify: `frontend/src/pages/profile/index.vue`
  复用真实历史数据和统一状态。

---

### Task 1: 建立真实用户边界并修复登录落库

**Files:**
- Create: `backend/internal/domain/user/service.go`
- Create: `backend/internal/domain/user/service_test.go`
- Create: `backend/internal/repository/postgres/user_repo.go`
- Modify: `backend/internal/http/handlers/auth_handler.go`
- Modify: `backend/internal/http/handlers/auth_handler_test.go`
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: 先写用户服务的失败测试**

```go
func TestGetOrCreateByMockCodeCreatesUserOnFirstLogin(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)

	user, created, err := svc.GetOrCreateByMockCode(context.Background(), "wx-code-1")

	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, "mock-openid-wx-code-1", user.OpenID)
	require.Equal(t, 3, user.FreeQuota)
}

func TestGetOrCreateByMockCodeReusesExistingUser(t *testing.T) {
	repo := newMockUserRepo()
	repo.usersByOpenID["mock-openid-wx-code-1"] = &User{
		ID:        42,
		OpenID:    "mock-openid-wx-code-1",
		Balance:   0,
		FreeQuota: 2,
	}
	svc := NewService(repo)

	user, created, err := svc.GetOrCreateByMockCode(context.Background(), "wx-code-1")

	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, int64(42), user.ID)
	require.Equal(t, 2, user.FreeQuota)
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./internal/domain/user -run TestGetOrCreateByMockCode -v`
Expected: FAIL，提示 `package image-play/internal/domain/user is not in std` 或 `undefined: NewService`

- [ ] **Step 3: 写登录 handler 的失败测试，要求返回真实落库用户**

```go
func TestLoginReturnsPersistedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := user.NewService(newMockUserRepo())
	r := gin.New()
	r.POST("/api/auth/login", LoginHandler("test-secret", svc))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"code":"wx-code-1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp LoginResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.AccessToken)
	require.Equal(t, "mock-openid-wx-code-1", resp.User.Openid)
	require.Equal(t, int64(3), resp.User.FreeQuota)
}
```

- [ ] **Step 4: 跑 handler 测试确认失败**

Run: `cd backend && go test ./internal/http/handlers -run TestLoginReturnsPersistedUser -v`
Expected: FAIL，提示 `too many arguments in call to LoginHandler` 或 `undefined: user.NewService`

- [ ] **Step 5: 写最小实现，让登录基于真实用户边界工作**

```go
type User struct {
	ID        int64
	OpenID    string
	Balance   float64
	FreeQuota int
}

type Repository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByOpenID(ctx context.Context, openID string) (*User, error)
	Create(ctx context.Context, user *User) error
}

func (s *Service) GetOrCreateByMockCode(ctx context.Context, code string) (*User, bool, error) {
	openID := "mock-openid-" + code
	existing, err := s.repo.GetByOpenID(ctx, openID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	user := &User{
		OpenID:    openID,
		Balance:   0,
		FreeQuota: 3,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, false, err
	}
	return user, true, nil
}
```

```go
func LoginHandler(jwtSecret string, userSvc *user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		account, _, err := userSvc.GetOrCreateByMockCode(c.Request.Context(), req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": strconv.FormatInt(account.ID, 10),
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
		})
```

- [ ] **Step 6: 跑相关测试确认通过**

Run: `cd backend && go test ./internal/domain/user ./internal/http/handlers -run 'Test(GetOrCreateByMockCode|LoginReturnsPersistedUser)' -v`
Expected: PASS

- [ ] **Step 7: 接上 `/me` 的真实用户读取并跑完整 handler 测试**

```go
func MeHandler(userRepo user.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64("user_id")
		account, err := userRepo.GetByID(c.Request.Context(), uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
			return
		}
		c.JSON(http.StatusOK, User{
			ID:        account.ID,
			Openid:    account.OpenID,
			Balance:   int64(account.Balance),
			FreeQuota: int64(account.FreeQuota),
		})
	}
}
```

Run: `cd backend && go test ./internal/http/handlers -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add backend/internal/domain/user backend/internal/repository/postgres/user_repo.go backend/internal/http/handlers/auth_handler.go backend/internal/http/handlers/auth_handler_test.go backend/internal/http/router.go
git commit -m "feat: persist mock login users"
```

### Task 2: 让客户端配置与模板列表只来自真实数据

**Files:**
- Create: `backend/internal/repository/postgres/scene_template_repo.go`
- Create: `backend/internal/http/handlers/templates_handler.go`
- Create: `backend/internal/http/handlers/templates_handler_test.go`
- Modify: `backend/internal/domain/scenes/catalog.go`
- Modify: `backend/internal/domain/scenes/templates.go`
- Modify: `backend/internal/http/handlers/config_handler.go`
- Modify: `backend/internal/http/router.go`
- Modify: `infra/sql/seed_scene_templates.sql`

- [ ] **Step 1: 先写模板查询 handler 的失败测试**

```go
func TestListSceneTemplatesReturnsOnlyActiveTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockSceneTemplateRepo{
		templates: []scenes.Template{
			{SceneKey: "portrait", TemplateKey: "office-pro", Name: "通勤职业", Active: true},
			{SceneKey: "portrait", TemplateKey: "disabled", Name: "停用模板", Active: false},
		},
	}
	r := gin.New()
	r.GET("/api/scenes/:scene_key/templates", ListSceneTemplatesHandler(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/scenes/portrait/templates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"items":[{"key":"office-pro","name":"通勤职业","scene_key":"portrait"}]}`, w.Body.String())
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./internal/http/handlers -run TestListSceneTemplatesReturnsOnlyActiveTemplates -v`
Expected: FAIL，提示 `undefined: ListSceneTemplatesHandler`

- [ ] **Step 3: 写客户端配置的失败测试，要求返回真实五个场景**

```go
func TestClientConfigReturnsSceneHallOrder(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ClientConfigHandler(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{
		"brand_slogan":"Play with your images",
		"pricing":{"single":"1.00","pack10":"8.00"},
		"scene_order":["portrait","festival","invitation","tshirt","poster"]
	}`, w.Body.String())
}
```

- [ ] **Step 4: 跑配置测试确认失败**

Run: `cd backend && go test ./internal/http/handlers -run TestClientConfigReturnsSceneHallOrder -v`
Expected: FAIL，scene_order 仍为 `portrait/landscape/fun`

- [ ] **Step 5: 写最小实现，建立模板 repo 和真实配置**

```go
type Template struct {
	ID             int64
	SceneKey       string
	TemplateKey    string
	Name           string
	FormSchema     []FormField
	PromptPreset   PromptPreset
	SampleImageURL string
	Active         bool
}

func SupportedSceneOrder() []string {
	return []string{
		ScenePortrait,
		SceneFestival,
		SceneInvitation,
		SceneTshirt,
		ScenePoster,
	}
}
```

```go
func ListSceneTemplatesHandler(repo SceneTemplateLister) gin.HandlerFunc {
	return func(c *gin.Context) {
		sceneKey := c.Param("scene_key")
		items, err := repo.ListActiveByScene(c.Request.Context(), sceneKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list templates"})
			return
		}

		resp := make([]gin.H, 0, len(items))
		for _, item := range items {
			resp = append(resp, gin.H{
				"key":              item.TemplateKey,
				"name":             item.Name,
				"scene_key":        item.SceneKey,
				"form_schema":      item.FormSchema,
				"sample_image_url": item.SampleImageURL,
			})
		}
		c.JSON(http.StatusOK, gin.H{"items": resp})
	}
}
```

```sql
INSERT INTO scene_templates (scene_key, template_key, name, form_schema, prompt_preset, sample_image_url)
VALUES
  ('portrait', 'office-pro', '通勤职业', '[{"name":"subject_name","label":"拍摄对象","type":"text","required":true}]', '{"base_prompt":"职业形象照","style_words":["professional","business"]}', 'https://example.com/portrait-office-pro.png'),
  ('festival', 'spring-festival', '春节祝福', '[{"name":"title","label":"标题","type":"text","required":true}]', '{"base_prompt":"春节贺卡","style_words":["festive","red and gold"]}', 'https://example.com/festival-spring.png'),
  ('invitation', 'wedding-classic', '婚礼请柬', '[{"name":"host_name","label":"主办人","type":"text","required":true}]', '{"base_prompt":"婚礼请柬","style_words":["elegant","romantic"]}', 'https://example.com/invitation-wedding.png'),
  ('tshirt', 'streetwear', '街头潮流', '[{"name":"theme","label":"主题","type":"text","required":true}]', '{"base_prompt":"街头潮流T恤图案","style_words":["streetwear","graffiti"]}', 'https://example.com/tshirt-streetwear.png'),
  ('poster', 'concert', '演唱会海报', '[{"name":"title","label":"标题","type":"text","required":true}]', '{"base_prompt":"演唱会海报","style_words":["concert","neon"]}', 'https://example.com/poster-concert.png')
ON CONFLICT (scene_key, template_key) DO UPDATE
SET name = EXCLUDED.name,
    form_schema = EXCLUDED.form_schema,
    prompt_preset = EXCLUDED.prompt_preset,
    sample_image_url = EXCLUDED.sample_image_url,
    is_active = true;
```

- [ ] **Step 6: 跑模板与配置测试确认通过**

Run: `cd backend && go test ./internal/http/handlers -run 'Test(ListSceneTemplatesReturnsOnlyActiveTemplates|ClientConfigReturnsSceneHallOrder)' -v`
Expected: PASS

- [ ] **Step 7: 把模板接口接进 router 并跑相关包测试**

```go
templateRepo := postgres.NewSceneTemplateRepo(db)
r.GET("/api/scenes/:scene_key/templates", handlers.ListSceneTemplatesHandler(templateRepo))
genSvc := generation.NewService(genRepo, templateRepo)
```

Run: `cd backend && go test ./internal/http ./internal/http/handlers ./internal/domain/scenes -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add backend/internal/repository/postgres/scene_template_repo.go backend/internal/http/handlers/templates_handler.go backend/internal/http/handlers/templates_handler_test.go backend/internal/domain/scenes/catalog.go backend/internal/domain/scenes/templates.go backend/internal/http/handlers/config_handler.go backend/internal/http/router.go infra/sql/seed_scene_templates.sql
git commit -m "feat: serve real scene templates and config"
```

### Task 3: 统一模板校验、Prompt Builder 和生成状态

**Files:**
- Modify: `backend/internal/domain/scenes/prompt_builder.go`
- Modify: `backend/internal/domain/scenes/prompt_builder_test.go`
- Modify: `backend/internal/domain/generation/service.go`
- Modify: `backend/internal/domain/generation/service_test.go`
- Modify: `backend/internal/domain/generation/memory_repo_test.go`
- Modify: `backend/internal/http/handlers/generations_handler_test.go`
- Modify: `backend/internal/worker/jobs/generation_job.go`
- Create: `backend/internal/worker/jobs/generation_job_test.go`
- Modify: `backend/internal/http/handlers/admin_metrics_handler.go`
- Modify: `backend/internal/http/handlers/admin_metrics_handler_test.go`

- [ ] **Step 1: 先写 generation service 的失败测试，要求禁用模板不能创建**

```go
func TestCreateGenerationRejectsInactiveTemplate(t *testing.T) {
	repo := newInMemoryRepo()
	templates := &mockTemplateLookup{
		template: nil,
	}
	svc := NewService(repo, templates)

	_, err := svc.CreateGeneration(context.Background(), CreateGenerationInput{
		UserID:          1,
		ClientRequestID: "req-1",
		SceneKey:        "portrait",
		TemplateKey:     "inactive-template",
		Fields:          map[string]string{"subject_name": "Alice"},
	})

	require.ErrorIs(t, err, ErrTemplateNotAvailable)
}
```

- [ ] **Step 2: 跑 generation service 测试确认失败**

Run: `cd backend && go test ./internal/domain/generation -run TestCreateGenerationRejectsInactiveTemplate -v`
Expected: FAIL，提示 `undefined: ErrTemplateNotAvailable` 或 `too many arguments in call to NewService`

- [ ] **Step 3: 写 worker 的失败测试，要求使用统一 Prompt Builder 并写入 success**

```go
func TestExecuteBuildsPromptFromTemplatePreset(t *testing.T) {
	repo := newJobRepo()
	templateRepo := &mockJobTemplateRepo{
		template: &scenes.Template{
			SceneKey:    "invitation",
			TemplateKey: "wedding-classic",
			PromptPreset: scenes.PromptPreset{
				BasePrompt: "婚礼请柬",
				StyleWords: []string{"elegant"},
			},
		},
	}
	model := &stubModelClient{}
	audit := &stubAuditClient{pass: true}
	job := NewGenerationJob(repo, templateRepo, model, audit, nil)

	err := job.Execute(context.Background(), &generation.Generation{
		ID:          1,
		UserID:      1,
		SceneKey:    "invitation",
		TemplateKey: "wedding-classic",
		Fields:      map[string]string{"host_name": "林然与苏晴"},
	})

	require.NoError(t, err)
	require.Contains(t, model.lastPrompt, "婚礼请柬")
	require.Contains(t, model.lastPrompt, "林然与苏晴")
	require.Equal(t, "success", repo.updatedStatus[1])
}
```

- [ ] **Step 4: 跑 worker 测试确认失败**

Run: `cd backend && go test ./internal/worker/jobs -run TestExecuteBuildsPromptFromTemplatePreset -v`
Expected: FAIL，提示 `too many arguments in call to NewGenerationJob` 或 `undefined: scenes.PromptPreset`

- [ ] **Step 5: 写 admin metrics 的失败测试，要求成功口径统计 `success`**

```go
func TestDashboardMetricsCountsSuccessGenerations(t *testing.T) {
	db := setupMetricsTestDB(t)
	_, err := db.Exec(`
		INSERT INTO generations (user_id, client_request_id, scene_key, template_key, fields, status, created_at, updated_at)
		VALUES (1, 'req-1', 'portrait', 'office-pro', '{}', 'success', NOW(), NOW())
	`)
	require.NoError(t, err)

	w := performMetricsRequest(t, db)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"portrait":1`)
}
```

- [ ] **Step 6: 跑 metrics 测试确认失败**

Run: `cd backend && go test ./internal/http/handlers -run TestDashboardMetricsCountsSuccessGenerations -v`
Expected: FAIL，因为 handler 仍统计 `completed`

- [ ] **Step 7: 写最小实现，统一模板校验、prompt 构建与状态口径**

```go
var ErrTemplateNotAvailable = errors.New("template not available")

type TemplateLookup interface {
	GetActiveTemplate(ctx context.Context, sceneKey, templateKey string) (*scenes.Template, error)
}

func (s *Service) CreateGeneration(ctx context.Context, input CreateGenerationInput) (int64, error) {
	template, err := s.templates.GetActiveTemplate(ctx, input.SceneKey, input.TemplateKey)
	if err != nil {
		return 0, err
	}
	if template == nil {
		return 0, ErrTemplateNotAvailable
	}
```

```go
func BuildPrompt(preset PromptPreset, fields map[string]string) string {
	parts := []string{preset.BasePrompt}
	if len(preset.StyleWords) > 0 {
		parts = append(parts, "风格关键词："+strings.Join(preset.StyleWords, ", "))
	}
```

```go
type TemplateLoader interface {
	GetActiveTemplate(ctx context.Context, sceneKey, templateKey string) (*scenes.Template, error)
}

func (j *GenerationJob) Execute(ctx context.Context, g *generation.Generation) error {
	template, err := j.templateRepo.GetActiveTemplate(ctx, g.SceneKey, g.TemplateKey)
	if err != nil {
		return err
	}
	if template == nil {
		_ = j.generationRepo.UpdateResult(ctx, g.ID, "failed", "")
		return fmt.Errorf("template not available")
	}

	prompt := scenes.BuildPrompt(template.PromptPreset, g.Fields)
```

```go
SELECT scene_key, COUNT(*)
FROM generations
WHERE status = 'success'
GROUP BY scene_key
```

- [ ] **Step 8: 跑相关测试确认通过**

Run: `cd backend && go test ./internal/domain/scenes ./internal/domain/generation ./internal/worker/jobs ./internal/http/handlers -run 'Test(CreateGenerationRejectsInactiveTemplate|ExecuteBuildsPromptFromTemplatePreset|DashboardMetricsCountsSuccessGenerations)' -v`
Expected: PASS

- [ ] **Step 9: 跑完整后端测试**

Run: `cd backend && go test ./...`
Expected: PASS

- [ ] **Step 10: Commit**

```bash
git add backend/internal/domain/scenes/prompt_builder.go backend/internal/domain/scenes/prompt_builder_test.go backend/internal/domain/generation/service.go backend/internal/domain/generation/service_test.go backend/internal/domain/generation/memory_repo_test.go backend/internal/http/handlers/generations_handler_test.go backend/internal/worker/jobs/generation_job.go backend/internal/worker/jobs/generation_job_test.go backend/internal/http/handlers/admin_metrics_handler.go backend/internal/http/handlers/admin_metrics_handler_test.go
git commit -m "feat: validate templates and unify generation status"
```

### Task 4: 修复前端运行骨架与 mock 会话初始化

**Files:**
- Create: `frontend/src/pages.json`
- Create: `frontend/src/services/session.ts`
- Create: `frontend/src/services/__tests__/api.test.ts`
- Modify: `frontend/src/main.ts`
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/services/api.ts`

- [ ] **Step 1: 安装前端依赖**

Run: `cd frontend && npm install`
Expected: 成功安装 `vue-tsc`、`vitest` 等依赖

- [ ] **Step 2: 先写 API 映射的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { mapSceneTemplate, mapHistoryItem } from '../api'

describe('api mapping', () => {
  it('maps scene template dto to frontend template', () => {
    expect(mapSceneTemplate({
      key: 'office-pro',
      name: '通勤职业',
      scene_key: 'portrait',
      form_schema: [{ name: 'subject_name', label: '拍摄对象', type: 'text', required: true }],
      sample_image_url: 'https://example.com/template.png',
    })).toEqual({
      key: 'office-pro',
      name: '通勤职业',
      sceneKey: 'portrait',
      formSchema: [{ name: 'subject_name', label: '拍摄对象', type: 'text', required: true }],
      sampleImageUrl: 'https://example.com/template.png',
    })
  })
})
```

- [ ] **Step 3: 跑测试确认失败**

Run: `cd frontend && npm test -- src/services/__tests__/api.test.ts --run`
Expected: FAIL，提示 `mapSceneTemplate is not exported`

- [ ] **Step 4: 写最小实现，补齐页面骨架、Pinia 和 mock 会话**

```ts
import { createSSRApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'

export function createApp() {
  const app = createSSRApp(App)
  app.use(createPinia())
  return { app }
}
```

```json
{
  "pages": [
    { "path": "pages/home/index", "style": { "navigationBarTitleText": "Image Play" } },
    { "path": "pages/scene/index", "style": { "navigationBarTitleText": "选择模板" } },
    { "path": "pages/result/index", "style": { "navigationBarTitleText": "生成结果" } },
    { "path": "pages/history/index", "style": { "navigationBarTitleText": "历史记录" } },
    { "path": "pages/profile/index", "style": { "navigationBarTitleText": "个人中心" } }
  ]
}
```

```ts
const MOCK_CODE_KEY = 'mock_login_code'

export async function ensureMockSession() {
  const existing = uni.getStorageSync('access_token')
  if (existing) return existing

  let code = uni.getStorageSync(MOCK_CODE_KEY)
  if (!code) {
    code = `mock-code-${Date.now()}`
    uni.setStorageSync(MOCK_CODE_KEY, code)
  }

  const resp = await login(code)
  uni.setStorageSync('access_token', resp.access_token)
  return resp.access_token
}
```

```vue
<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { ensureMockSession } from './services/session'

onLaunch(() => {
  void ensureMockSession()
})
</script>
```

```ts
export function mapSceneTemplate(dto: SceneTemplateDTO): Template {
  return {
    key: dto.key,
    name: dto.name,
    sceneKey: dto.scene_key,
    formSchema: dto.form_schema,
    sampleImageUrl: dto.sample_image_url,
  }
}
```

- [ ] **Step 5: 把 401 回跳改到真实存在的首页并新增模板/生成接口**

```ts
if (response.statusCode === 401) {
  uni.removeStorageSync('access_token')
  uni.reLaunch({ url: '/pages/home/index' })
  reject(new Error('Unauthorized'))
  return
}

export function getSceneTemplates(sceneKey: string) {
  return request<{ items: SceneTemplateDTO[] }>({
    url: `/api/scenes/${sceneKey}/templates`,
    method: 'GET',
  }).then((res) => ({
    items: res.items.map(mapSceneTemplate),
  }))
}

export function createGeneration(payload: CreateGenerationPayload) {
  return request<{ generation_id: number }>({
    url: '/api/generations',
    method: 'POST',
    data: payload,
    headers: { 'Content-Type': 'application/json' },
  })
}
```

- [ ] **Step 6: 跑 API 测试与类型检查确认通过**

Run: `cd frontend && npm test -- src/services/__tests__/api.test.ts --run && npm run type-check`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages.json frontend/src/main.ts frontend/src/App.vue frontend/src/services/session.ts frontend/src/services/api.ts frontend/src/services/__tests__/api.test.ts
git commit -m "feat: bootstrap frontend session and runtime skeleton"
```

### Task 5: 接通前端场景、生成、历史和结果页真实流程

**Files:**
- Create: `frontend/src/utils/generation.ts`
- Create: `frontend/src/utils/__tests__/generation.test.ts`
- Modify: `frontend/src/store/generation.ts`
- Modify: `frontend/src/store/__tests__/generation.test.ts`
- Modify: `frontend/src/pages/home/index.vue`
- Modify: `frontend/src/pages/scene/index.vue`
- Modify: `frontend/src/pages/result/index.vue`
- Modify: `frontend/src/pages/history/index.vue`
- Modify: `frontend/src/pages/profile/index.vue`

- [ ] **Step 1: 先写 generation store 的失败测试**

```ts
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useGenerationStore } from '../generation'
import * as api from '../../services/api'

vi.mock('../../services/api', () => ({
  getSceneTemplates: vi.fn(),
  createGeneration: vi.fn(),
}))

describe('generation store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('loads templates for a scene', async () => {
    vi.mocked(api.getSceneTemplates).mockResolvedValue({
      items: [{ key: 'office-pro', name: '通勤职业', sceneKey: 'portrait', formSchema: [] }],
    })
    const store = useGenerationStore()

    await store.loadTemplates('portrait')

    expect(store.selectedScene).toBe('portrait')
    expect(store.templates).toHaveLength(1)
  })

  it('submits generation and stores generation id', async () => {
    vi.mocked(api.createGeneration).mockResolvedValue({ generation_id: 88 })
    const store = useGenerationStore()
    store.selectedScene = 'portrait'
    store.selectedTemplate = { key: 'office-pro', name: '通勤职业', sceneKey: 'portrait', formSchema: [] }

    await store.submitGeneration()

    expect(store.lastGenerationId).toBe(88)
    expect(store.isSubmitting).toBe(false)
  })
})
```

- [ ] **Step 2: 跑 store 测试确认失败**

Run: `cd frontend && npm test -- src/store/__tests__/generation.test.ts --run`
Expected: FAIL，提示 `loadTemplates is not a function` 或 `lastGenerationId does not exist`

- [ ] **Step 3: 写历史纯逻辑的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { filterHistoryItems, findHistoryItemById, isGenerationPending } from '../generation'

describe('generation utils', () => {
  const items = [
    { id: 1, status: 'queued', resultUrl: '' },
    { id: 2, status: 'success', resultUrl: 'https://example.com/result.png' },
  ]

  it('filters by unified status', () => {
    expect(filterHistoryItems(items, 'success')).toEqual([items[1]])
  })

  it('finds a history item by generation id', () => {
    expect(findHistoryItemById(items, 2)?.status).toBe('success')
  })

  it('treats queued as pending', () => {
    expect(isGenerationPending('queued')).toBe(true)
  })
})
```

- [ ] **Step 4: 跑 utils 测试确认失败**

Run: `cd frontend && npm test -- src/utils/__tests__/generation.test.ts --run`
Expected: FAIL，提示 `Cannot find module '../generation'`

- [ ] **Step 5: 写最小实现，接通真实模板与真实提交流程**

```ts
export const useGenerationStore = defineStore('generation', {
  state: () => ({
    selectedScene: '',
    templates: [] as Template[],
    selectedTemplate: null as Template | null,
    formValues: {} as Record<string, string>,
    isSubmitting: false,
    submitError: '',
    lastGenerationId: null as number | null,
  }),
  actions: {
    async loadTemplates(sceneKey: string) {
      const res = await getSceneTemplates(sceneKey)
      this.selectedScene = sceneKey
      this.templates = res.items
      this.selectedTemplate = res.items[0] ?? null
      this.formValues = {}
    },
    async submitGeneration() {
      if (!this.selectedScene || !this.selectedTemplate) {
        throw new Error('missing scene or template')
      }
      this.isSubmitting = true
      this.submitError = ''
      try {
        const res = await createGeneration({
          client_request_id: `${Date.now()}`,
          scene_key: this.selectedScene,
          template_key: this.selectedTemplate.key,
          fields: this.formValues,
        })
        this.lastGenerationId = res.generation_id
        return res.generation_id
      } catch (error) {
        this.submitError = error instanceof Error ? error.message : '提交失败'
        throw error
      } finally {
        this.isSubmitting = false
      }
    },
  },
})
```

```ts
export function isGenerationPending(status: string) {
  return ['queued', 'running', 'result_auditing'].includes(status)
}

export function filterHistoryItems<T extends { status: string }>(items: T[], status: string) {
  if (status === 'all') return items
  return items.filter((item) => item.status === status)
}

export function findHistoryItemById<T extends { id: number }>(items: T[], id: number) {
  return items.find((item) => item.id === id)
}
```

```vue
onMounted(async () => {
  await generationStore.loadTemplates(sceneKey.value)
})

async function handleSubmit() {
  const generationId = await generationStore.submitGeneration()
  uni.navigateTo({ url: `/pages/result/index?generation_id=${generationId}` })
}
```

- [ ] **Step 6: 修正首页/历史/结果页逻辑**

```vue
<view v-if="sceneOrder.length > 0">
  ...
</view>
<view v-else-if="error">{{ error }}</view>
<view v-else class="loading">
  <text>加载中...</text>
</view>
```

```vue
const filters = ['all', 'queued', 'running', 'result_auditing', 'success', 'failed']
const filteredItems = computed(() => filterHistoryItems(items.value, filter.value))

function goToResult(item: HistoryItem) {
  uni.navigateTo({ url: `/pages/result/index?generation_id=${item.id}` })
}
```

```vue
async function refreshHistory() {
  const res = await getHistory()
  historyItems.value = res.items.map(mapHistoryItem)
  currentItem.value = findHistoryItemById(historyItems.value, Number(generationId.value)) ?? null
}

watchEffect(() => {
  if (currentItem.value && isGenerationPending(currentItem.value.status)) {
    timer = setTimeout(() => { void refreshHistory() }, 1500)
  }
})
```

- [ ] **Step 7: 跑前端测试与类型检查**

Run: `cd frontend && npm test -- src/store/__tests__/generation.test.ts src/utils/__tests__/generation.test.ts --run && npm run type-check`
Expected: PASS

- [ ] **Step 8: 手工跑一遍主页面流程**

Run: `cd frontend && npm run dev`
Expected: 首页可进入，场景页能拉模板，提交后跳结果页，结果页在任务完成后显示图片或失败状态

- [ ] **Step 9: Commit**

```bash
git add frontend/src/store/generation.ts frontend/src/store/__tests__/generation.test.ts frontend/src/utils/generation.ts frontend/src/utils/__tests__/generation.test.ts frontend/src/pages/home/index.vue frontend/src/pages/scene/index.vue frontend/src/pages/result/index.vue frontend/src/pages/history/index.vue frontend/src/pages/profile/index.vue
git commit -m "feat: connect frontend scene hall flow to real APIs"
```

### Task 6: 全量验证并更新项目状态

**Files:**
- Modify: `PROGRESS.md`

- [ ] **Step 1: 跑后端完整测试**

Run: `cd backend && go test ./...`
Expected: PASS

- [ ] **Step 2: 跑前端测试与类型检查**

Run: `cd frontend && npm test -- --run && npm run type-check`
Expected: PASS

- [ ] **Step 3: 更新进度文档，只保留真实验证过的状态**

```md
## 当前验证状态

- 后端编译：通过（api + worker）
- 后端单元测试：通过
- 前端单元测试：通过
- 前端类型检查：通过
- 模板查询/生成/历史主链路：已联调
- 支付与后台权限：待后续批次修复
```

- [ ] **Step 4: 跑工作区检查**

Run: `git status --short`
Expected: 只剩本次修复相关文件改动

- [ ] **Step 5: Commit**

```bash
git add PROGRESS.md
git commit -m "docs: update boundary repair verification status"
```

---

## Self-Review

### Spec Coverage

- 用户真实落库：Task 1
- 客户端配置与五个场景一致：Task 2
- 模板真实下发：Task 2
- 生成前模板校验：Task 3
- worker 使用统一 Prompt Builder：Task 3
- 状态统一：Task 3 + Task 5
- 前端真实提交流程：Task 4 + Task 5
- 历史/结果页基于真实任务：Task 5
- 测试与验证：Task 1-6

### Placeholder Scan

- 无 `TODO` / `TBD`
- 每个代码步骤都包含了明确的测试或实现片段
- 每个任务都包含命令与预期输出

### Type Consistency

- 后端模板字段统一使用 `SceneKey / TemplateKey / PromptPreset`
- 前端模板字段统一映射为 `sceneKey / formSchema / sampleImageUrl`
- 生成状态统一为 `queued / running / result_auditing / success / failed`
