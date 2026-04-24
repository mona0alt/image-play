# AI 图片场景馆 MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付一个可上线验证的微信小程序 MVP，包含 5 个场景馆入口、模板驱动生成、支付计费、历史记录、埋点分析和基础后台能力。

**Architecture:** 采用 `UniApp + Vue3` 前端、`Go API + Go Worker + PostgreSQL` 后端的异步任务架构。前端只负责场景选择、表单提交和结果展示；后端负责鉴权、模板配置、资产上传、任务调度、审核、扣费和统计埋点。

**Tech Stack:** UniApp、Vue 3、TypeScript、Vitest、Go 1.22、Gin、sqlc 或 GORM、PostgreSQL 15、Docker Compose、腾讯云 COS、微信支付、Higress

---

## File Structure

### Frontend

- Create: `frontend/package.json`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/App.vue`
- Create: `frontend/src/pages/home/index.vue`
- Create: `frontend/src/pages/scene/index.vue`
- Create: `frontend/src/pages/result/index.vue`
- Create: `frontend/src/pages/history/index.vue`
- Create: `frontend/src/pages/profile/index.vue`
- Create: `frontend/src/components/scene/SceneHeroCard.vue`
- Create: `frontend/src/components/scene/SceneGalleryCard.vue`
- Create: `frontend/src/components/scene/TemplatePicker.vue`
- Create: `frontend/src/components/form/SceneFieldForm.vue`
- Create: `frontend/src/components/result/ResultPreviewCard.vue`
- Create: `frontend/src/store/user.ts`
- Create: `frontend/src/store/config.ts`
- Create: `frontend/src/store/generation.ts`
- Create: `frontend/src/services/api.ts`
- Create: `frontend/src/services/tracking.ts`
- Create: `frontend/src/types/scene.ts`
- Test: `frontend/src/services/__tests__/tracking.test.ts`
- Test: `frontend/src/store/__tests__/generation.test.ts`

### Backend API

- Create: `backend/go.mod`
- Create: `backend/cmd/api/main.go`
- Create: `backend/cmd/worker/main.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/http/router.go`
- Create: `backend/internal/http/middleware/auth.go`
- Create: `backend/internal/http/handlers/auth_handler.go`
- Create: `backend/internal/http/handlers/config_handler.go`
- Create: `backend/internal/http/handlers/assets_handler.go`
- Create: `backend/internal/http/handlers/generations_handler.go`
- Create: `backend/internal/http/handlers/payments_handler.go`
- Create: `backend/internal/http/handlers/history_handler.go`
- Create: `backend/internal/domain/scenes/catalog.go`
- Create: `backend/internal/domain/scenes/templates.go`
- Create: `backend/internal/domain/scenes/prompt_builder.go`
- Create: `backend/internal/domain/billing/service.go`
- Create: `backend/internal/domain/generation/service.go`
- Create: `backend/internal/domain/assets/service.go`
- Create: `backend/internal/domain/tracking/service.go`
- Create: `backend/internal/worker/runner.go`
- Create: `backend/internal/worker/jobs/generation_job.go`
- Create: `backend/internal/repository/postgres/*.go`
- Test: `backend/internal/domain/scenes/prompt_builder_test.go`
- Test: `backend/internal/domain/billing/service_test.go`
- Test: `backend/internal/domain/generation/service_test.go`
- Test: `backend/internal/http/handlers/generations_handler_test.go`

### Database / Infra / Docs

- Create: `backend/migrations/0001_init.sql`
- Create: `backend/migrations/0002_scene_templates.sql`
- Create: `backend/migrations/0003_tracking_events.sql`
- Create: `infra/docker-compose.yml`
- Create: `infra/env/api.env.example`
- Create: `infra/env/worker.env.example`
- Create: `infra/sql/seed_scene_templates.sql`
- Create: `docs/ops/mvp-runbook.md`
- Create: `docs/ops/scene-hall-metrics.md`

---

### Task 1: 初始化仓库结构与本地开发骨架

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/App.vue`
- Create: `backend/go.mod`
- Create: `backend/cmd/api/main.go`
- Create: `backend/cmd/worker/main.go`
- Create: `infra/docker-compose.yml`
- Create: `infra/env/api.env.example`
- Create: `infra/env/worker.env.example`

- [ ] **Step 1: 写一个最小化的前后端启动测试清单**

```text
前端: npm run dev 能启动空白 UniApp 页面
后端 API: go run ./cmd/api 能返回 /healthz 200
后端 Worker: go run ./cmd/worker 能输出 "worker started"
数据库: docker compose up postgres 后能接受连接
```

- [ ] **Step 2: 建立前端脚手架并验证能启动**

Run: `cd frontend && npm install`
Expected: 成功安装依赖，无 peer dependency 阻塞错误

- [ ] **Step 3: 建立后端 go module 和健康检查**

```go
func main() {
    r := gin.Default()
    r.GET("/healthz", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    _ = r.Run(":8080")
}
```

- [ ] **Step 4: 建立 worker 启动骨架**

```go
func main() {
    log.Println("worker started")
}
```

- [ ] **Step 5: 加入 Docker Compose 最小运行环境**

```yaml
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: image_play
    ports:
      - "5432:5432"
```

- [ ] **Step 6: 验证骨架可以本地跑通**

Run: `cd backend && go test ./...`
Expected: `ok` 或 `?[no test files]`

- [ ] **Step 7: Commit**

```bash
git add frontend backend infra
git commit -m "chore: bootstrap scene hall mvp skeleton"
```

### Task 2: 建立数据库模型与迁移

**Files:**
- Create: `backend/migrations/0001_init.sql`
- Create: `backend/migrations/0002_scene_templates.sql`
- Create: `backend/migrations/0003_tracking_events.sql`
- Create: `infra/sql/seed_scene_templates.sql`
- Test: `backend/internal/domain/billing/service_test.go`

- [ ] **Step 1: 为核心表写迁移前的断言清单**

```sql
-- users, assets, generations, orders, transactions, system_configs,
-- scene_templates, tracking_events
```

- [ ] **Step 2: 写 `0001_init.sql` 创建核心业务表**

```sql
create table users (
  id bigserial primary key,
  openid varchar(64) not null unique,
  balance numeric(10,2) not null default 0,
  free_quota int not null default 3,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now()
);
```

- [ ] **Step 3: 写 `generations` 的幂等与并发约束**

```sql
create unique index idx_generations_user_request
on generations(user_id, client_request_id);

create unique index idx_generations_user_active
on generations(user_id)
where status in ('queued', 'running', 'result_auditing');
```

- [ ] **Step 4: 写场景模板和埋点表迁移**

```sql
create table scene_templates (
  id bigserial primary key,
  scene_key varchar(32) not null,
  template_key varchar(64) not null,
  name varchar(128) not null,
  form_schema jsonb not null,
  prompt_preset jsonb not null,
  sample_image_url varchar(255),
  is_active boolean not null default true
);
```

- [ ] **Step 5: 写模板种子数据**

```sql
insert into scene_templates (scene_key, template_key, name)
values ('portrait', 'office-pro', '通勤职业');
```

- [ ] **Step 6: 运行迁移并验证表结构**

Run: `docker compose -f infra/docker-compose.yml up -d postgres`
Expected: PostgreSQL 启动成功

- [ ] **Step 7: Commit**

```bash
git add backend/migrations infra/sql
git commit -m "feat: add initial schema and scene template seeds"
```

### Task 3: 实现场景目录、模板配置与 Prompt Builder

**Files:**
- Create: `backend/internal/domain/scenes/catalog.go`
- Create: `backend/internal/domain/scenes/templates.go`
- Create: `backend/internal/domain/scenes/prompt_builder.go`
- Test: `backend/internal/domain/scenes/prompt_builder_test.go`

- [ ] **Step 1: 先写 Prompt Builder 的失败测试**

```go
func TestBuildPromptForInvitation(t *testing.T) {
    input := BuildInput{
        SceneKey: "invitation",
        TemplateKey: "wedding-classic",
        Fields: map[string]string{
            "host_name": "林然与苏晴",
            "event_time": "2026-10-01 18:00",
            "event_place": "杭州西湖国宾馆",
        },
    }

    prompt := BuildPrompt(input)
    require.Contains(t, prompt, "婚礼请柬")
    require.Contains(t, prompt, "林然与苏晴")
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./internal/domain/scenes -run TestBuildPromptForInvitation -v`
Expected: FAIL，提示 `undefined: BuildPrompt`

- [ ] **Step 3: 实现场景目录常量和模板定义**

```go
const (
    ScenePortrait   = "portrait"
    SceneFestival   = "festival"
    SceneInvitation = "invitation"
    SceneTshirt     = "tshirt"
    ScenePoster     = "poster"
)
```

- [ ] **Step 4: 实现最小 Prompt Builder**

```go
func BuildPrompt(input BuildInput) string {
    return fmt.Sprintf("%s | %s | %+v", input.SceneKey, input.TemplateKey, input.Fields)
}
```

- [ ] **Step 5: 扩展为模板驱动文案拼装**

```go
type TemplatePreset struct {
    BasePrompt string
    StyleWords []string
}
```

- [ ] **Step 6: 跑测试确认通过并补更多模板测试**

Run: `cd backend && go test ./internal/domain/scenes -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/scenes
git commit -m "feat: add scene catalog and prompt builder"
```

### Task 4: 实现登录、用户信息与客户端配置接口

**Files:**
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/http/router.go`
- Create: `backend/internal/http/middleware/auth.go`
- Create: `backend/internal/http/handlers/auth_handler.go`
- Create: `backend/internal/http/handlers/config_handler.go`
- Create: `frontend/src/store/user.ts`
- Create: `frontend/src/store/config.ts`
- Create: `frontend/src/services/api.ts`
- Test: `backend/internal/http/handlers/auth_handler_test.go`

- [ ] **Step 1: 先写 `/api/auth/login` 的失败测试**

```go
func TestLoginReturnsTokenAndUser(t *testing.T) {
    reqBody := `{"code":"mock-wechat-code"}`
    // 断言返回 access_token 和 user 基础信息
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./internal/http/handlers -run TestLoginReturnsTokenAndUser -v`
Expected: FAIL，提示 handler 未注册

- [ ] **Step 3: 实现假登录流程和 JWT 签发**

```go
type LoginResponse struct {
    AccessToken string `json:"access_token"`
    User        UserDTO `json:"user"`
}
```

- [ ] **Step 4: 实现 `/api/me` 和 `/api/configs/client`**

```go
type ClientConfig struct {
    BrandSlogan string            `json:"brand_slogan"`
    Pricing     map[string]string `json:"pricing"`
    SceneOrder  []string          `json:"scene_order"`
}
```

- [ ] **Step 5: 前端实现用户与配置 store**

```ts
export const useUserStore = defineStore('user', {
  state: () => ({ token: '', profile: null as UserProfile | null })
})
```

- [ ] **Step 6: 运行前后端单测**

Run: `cd backend && go test ./internal/http/... -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal frontend/src/store frontend/src/services
git commit -m "feat: add auth and client config endpoints"
```

### Task 5: 实现首页场景馆、模板选择与表单渲染

**Files:**
- Create: `frontend/src/pages/home/index.vue`
- Create: `frontend/src/pages/scene/index.vue`
- Create: `frontend/src/components/scene/SceneHeroCard.vue`
- Create: `frontend/src/components/scene/SceneGalleryCard.vue`
- Create: `frontend/src/components/scene/TemplatePicker.vue`
- Create: `frontend/src/components/form/SceneFieldForm.vue`
- Create: `frontend/src/types/scene.ts`
- Test: `frontend/src/store/__tests__/generation.test.ts`

- [ ] **Step 1: 写首页渲染顺序的失败测试**

```ts
it('renders portrait as hero and remaining scenes as gallery cards', () => {
  const order = ['portrait', 'festival', 'invitation', 'tshirt', 'poster']
  expect(getHero(order)).toBe('portrait')
  expect(getGallery(order)).toEqual(['festival', 'invitation', 'tshirt', 'poster'])
})
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd frontend && npm test -- generation.test.ts`
Expected: FAIL，提示 `getHero` 未定义

- [ ] **Step 3: 实现首页艺廊布局和主推 Hero**

```vue
<SceneHeroCard :scene="heroScene" />
<SceneGalleryCard v-for="scene in galleryScenes" :key="scene.key" :scene="scene" />
```

- [ ] **Step 4: 实现场景页的模板驱动表单**

```vue
<TemplatePicker :templates="templates" v-model="selectedTemplate" />
<SceneFieldForm :schema="selectedTemplate.formSchema" v-model="formValues" />
```

- [ ] **Step 5: 把自由 Prompt 放到“高级设置”折叠区**

```vue
<Collapse title="高级设置（可选）">
  <textarea v-model="advancedPrompt" />
</Collapse>
```

- [ ] **Step 6: 跑前端测试和本地页面 smoke test**

Run: `cd frontend && npm test`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add frontend/src
git commit -m "feat: add scene hall home and template driven scene forms"
```

### Task 6: 实现 COS 上传、生成任务创建与 Worker 执行链路

**Files:**
- Create: `backend/internal/domain/assets/service.go`
- Create: `backend/internal/domain/generation/service.go`
- Create: `backend/internal/http/handlers/assets_handler.go`
- Create: `backend/internal/http/handlers/generations_handler.go`
- Create: `backend/internal/worker/runner.go`
- Create: `backend/internal/worker/jobs/generation_job.go`
- Create: `frontend/src/store/generation.ts`
- Test: `backend/internal/domain/generation/service_test.go`
- Test: `backend/internal/http/handlers/generations_handler_test.go`

- [ ] **Step 1: 先写“同用户只允许一个活跃任务”的失败测试**

```go
func TestCreateGenerationRejectsWhenActiveJobExists(t *testing.T) {
    err := svc.CreateGeneration(ctx, CreateGenerationInput{
        UserID: 1,
        ClientRequestID: "req-2",
    })
    require.ErrorIs(t, err, ErrActiveGenerationExists)
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./internal/domain/generation -run TestCreateGenerationRejectsWhenActiveJobExists -v`
Expected: FAIL，错误类型未定义

- [ ] **Step 3: 实现上传凭证与上传回执接口**

```go
type UploadIntentResponse struct {
    AssetID    int64  `json:"asset_id"`
    ObjectKey  string `json:"object_key"`
    UploadURL  string `json:"upload_url"`
    ExpireAt   string `json:"expire_at"`
}
```

- [ ] **Step 4: 实现生成任务创建接口**

```go
type CreateGenerationRequest struct {
    ClientRequestID string            `json:"client_request_id"`
    SceneKey        string            `json:"scene_key"`
    TemplateKey     string            `json:"template_key"`
    Fields          map[string]string `json:"fields"`
    SourceAssetID   *int64            `json:"source_asset_id,omitempty"`
}
```

- [ ] **Step 5: 实现 Worker 领取任务和状态流转**

```go
for {
    job, ok := repo.DequeueGeneration(ctx)
    if !ok {
        time.Sleep(time.Second)
        continue
    }
    _ = jobRunner.Run(ctx, job)
}
```

- [ ] **Step 6: 使用假模型客户端和假审核客户端跑通单元测试**

Run: `cd backend && go test ./internal/domain/generation ./internal/http/handlers -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal frontend/src/store
git commit -m "feat: add asset upload and async generation pipeline"
```

### Task 7: 实现计费、套餐支付与任务成功扣费

**Files:**
- Create: `backend/internal/domain/billing/service.go`
- Create: `backend/internal/http/handlers/payments_handler.go`
- Create: `backend/internal/http/handlers/history_handler.go`
- Create: `frontend/src/pages/profile/index.vue`
- Test: `backend/internal/domain/billing/service_test.go`

- [ ] **Step 1: 先写“任务成功只扣一次费”的失败测试**

```go
func TestChargeGenerationOnlyOnce(t *testing.T) {
    err1 := svc.ChargeGeneration(ctx, 1001)
    err2 := svc.ChargeGeneration(ctx, 1001)
    require.NoError(t, err1)
    require.ErrorIs(t, err2, ErrGenerationAlreadyCharged)
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./internal/domain/billing -run TestChargeGenerationOnlyOnce -v`
Expected: FAIL

- [ ] **Step 3: 实现套餐查询和订单创建**

```go
type PackageDTO struct {
    Code  string `json:"code"`
    Title string `json:"title"`
    Price string `json:"price"`
    Count int    `json:"count"`
}
```

- [ ] **Step 4: 实现支付回调幂等和余额入账**

```go
func (s *Service) HandlePaymentCallback(ctx context.Context, wxOrderNo string) error {
    // 幂等检查 -> 更新订单 -> 写 recharge/bonus 流水
    return nil
}
```

- [ ] **Step 5: 在 Worker 成功终态里接入扣费服务**

```go
if auditPassed {
    if err := billingSvc.ChargeGeneration(ctx, job.ID); err != nil { ... }
}
```

- [ ] **Step 6: 前端接入套餐页和剩余次数展示**

Run: `cd frontend && npm test`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add backend/internal/domain/billing backend/internal/http/handlers/payments_handler.go frontend/src/pages/profile
git commit -m "feat: add package billing and payment flow"
```

### Task 8: 实现结果页、历史记录、保存分享与埋点

**Files:**
- Create: `frontend/src/pages/result/index.vue`
- Create: `frontend/src/pages/history/index.vue`
- Create: `frontend/src/components/result/ResultPreviewCard.vue`
- Create: `frontend/src/services/tracking.ts`
- Create: `backend/internal/domain/tracking/service.go`
- Test: `frontend/src/services/__tests__/tracking.test.ts`

- [ ] **Step 1: 先写埋点事件映射的失败测试**

```ts
it('maps save and share actions to expected event names', () => {
  expect(mapTrackingEvent('save')).toBe('generation_saved')
  expect(mapTrackingEvent('share')).toBe('generation_shared')
})
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd frontend && npm test -- tracking.test.ts`
Expected: FAIL

- [ ] **Step 3: 实现前端埋点 SDK**

```ts
export function track(event: string, payload: Record<string, unknown>) {
  return api.post('/api/tracking/events', { event, payload })
}
```

- [ ] **Step 4: 实现结果页保存、分享、再来一张**

```vue
<ResultPreviewCard :image-url="resultUrl" @save="handleSave" @share="handleShare" />
```

- [ ] **Step 5: 实现历史记录列表和筛选**

```vue
<view v-for="item in historyList" :key="item.id">
  <text>{{ item.sceneName }}</text>
  <text>{{ item.status }}</text>
</view>
```

- [ ] **Step 6: 验证埋点落库**

Run: `cd backend && go test ./internal/domain/tracking -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/result frontend/src/pages/history frontend/src/services/tracking.ts backend/internal/domain/tracking
git commit -m "feat: add result history and analytics tracking"
```

### Task 9: 实现最小后台、运营指标与上线校验

**Files:**
- Create: `docs/ops/mvp-runbook.md`
- Create: `docs/ops/scene-hall-metrics.md`
- Create: `backend/internal/http/handlers/admin_metrics_handler.go`
- Create: `backend/internal/http/handlers/admin_templates_handler.go`
- Test: `backend/internal/http/handlers/admin_metrics_handler_test.go`

- [ ] **Step 1: 先写指标聚合接口的失败测试**

```go
func TestDashboardReturnsSceneConversionMetrics(t *testing.T) {
    // 断言返回 scene_clicks, generation_success, payments
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./internal/http/handlers -run TestDashboardReturnsSceneConversionMetrics -v`
Expected: FAIL

- [ ] **Step 3: 实现最小后台指标接口**

```go
type DashboardMetrics struct {
    SceneClicks       map[string]int `json:"scene_clicks"`
    GenerationSuccess map[string]int `json:"generation_success"`
    Payments          map[string]int `json:"payments"`
}
```

- [ ] **Step 4: 实现模板启停接口**

```go
type ToggleTemplateRequest struct {
    Active bool `json:"active"`
}
```

- [ ] **Step 5: 写上线 Runbook**

```markdown
1. 配置微信支付参数
2. 配置 COS 密钥
3. 运行迁移
4. 导入模板种子
5. 启动 API 与 Worker
6. 验证 healthz 与 metrics
```

- [ ] **Step 6: 运行最小验收测试**

Run: `cd backend && go test ./... && cd ../frontend && npm test`
Expected: 全部通过

- [ ] **Step 7: Commit**

```bash
git add backend/internal/http/handlers docs/ops
git commit -m "feat: add admin metrics and mvp runbook"
```

### Task 10: 端到端验收与灰度上线准备

**Files:**
- Modify: `docs/ops/mvp-runbook.md`
- Modify: `docs/ops/scene-hall-metrics.md`

- [ ] **Step 1: 按真实业务链路写验收清单**

```markdown
- 新用户登录后可看到 3 次免费额度
- 首页显示 5 个场景
- 头像写真可完整生成
- 请柬邀请可完整生成
- 支付完成后余额到账
- 生成成功后可保存、分享、进历史
```

- [ ] **Step 2: 跑本地联调**

Run: `docker compose -f infra/docker-compose.yml up -d`
Expected: postgres、api、worker 全部健康

- [ ] **Step 3: 执行手工冒烟测试**

Run: `docs/ops/mvp-runbook.md` 中逐项执行
Expected: 关键链路全部通过

- [ ] **Step 4: 检查监控指标已可观测**

Run: `curl http://localhost:8080/api/admin/dashboard`
Expected: 返回 200，且包含场景转化数据

- [ ] **Step 5: 完成首发前文档确认**

```markdown
- 套餐价格确认
- 模板样例图确认
- 审核文案确认
- 隐私政策确认
```

- [ ] **Step 6: Commit**

```bash
git add docs/ops
git commit -m "chore: finalize scene hall mvp release checklist"
```

---

## Plan Notes

- 先做统一任务链路，再填充 5 个场景模板，不要先做 5 套分叉流程。
- 首页视觉精致感主要来自样例图、排版、层级和留白，不来自复杂动效。
- 验证重点是场景差异化漏斗，不是总 DAU。
- `头像写真` 和 `请柬邀请` 应作为首批重点验收场景，其他场景只要流程一致即可复用。
