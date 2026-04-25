# AI Image Scene Hall MVP 上线 Runbook

## 前置检查清单

- [ ] 服务器环境已准备（Go 1.22+, PostgreSQL 14+, Redis 可选）
- [ ] 域名与 HTTPS 证书已配置
- [ ] 微信小程序 AppID 和 AppSecret 已配置

## 配置项

| 配置项 | 环境变量 | 说明 |
|--------|----------|------|
| 数据库连接 | `DATABASE_URL` | PostgreSQL 连接字符串 |
| JWT 密钥 | `JWT_SECRET` | 至少 32 字节随机字符串 |
| 微信支付 | `WXPAY_MCHID`, `WXPAY_APIV3_KEY` | 商户号和 APIv3 密钥 |
| 对象存储 | `COS_SECRET_ID`, `COS_SECRET_KEY`, `COS_BUCKET`, `COS_REGION` | 腾讯云 COS 配置 |
| AI 服务 | `AI_API_KEY`, `AI_BASE_URL` | 图像生成服务配置 |

## 上线步骤

### 1. 运行数据库迁移

```bash
cd backend
psql $DATABASE_URL -f migrations/0001_init.sql
psql $DATABASE_URL -f migrations/0002_scene_templates.sql
psql $DATABASE_URL -f migrations/0003_tracking_events.sql
```

### 2. 初始化场景模板

```bash
go run cmd/seed/main.go
```

确认 `scene_templates` 表中已插入默认模板数据。

### 3. 启动服务

```bash
# API 服务
go run cmd/api/main.go

# Worker 服务（建议单独部署）
go run cmd/worker/main.go
```

### 4. 健康检查

```bash
curl https://<your-domain>/healthz
# 期望返回: {"status":"ok"}
```

### 5. 验证后台指标接口

```bash
curl -H "Authorization: Bearer <token>" https://<your-domain>/api/admin/metrics
# 期望返回包含 scene_clicks, generation_success, payments 的 JSON
```

### 6. 验证模板启停接口

```bash
curl -X PUT -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"active":false}' \
  https://<your-domain>/api/admin/templates/1/toggle
```

## 回滚预案

- 数据库：保留迁移前的备份快照
- 服务：使用 systemd/docker 快速重启旧版本
- 支付：若支付回调异常，检查 `orders` 表状态并手动补单

## 端到端验收清单

按真实业务链路逐项验证：

- [ ] 新用户登录后，免费额度显示为 3 次
- [ ] 首页显示 5 个场景馆入口（头像写真、节日祝福、请柬邀请、T 恤设计、海报生成）
- [ ] 点击场景进入模板选择页，表单渲染正确
- [ ] 头像写真场景可完整提交生成任务
- [ ] 请柬邀请场景可完整提交生成任务
- [ ] 支付套餐后余额/次数正确到账
- [ ] 生成成功后结果页可保存图片、分享、点击再来一张
- [ ] 历史记录页展示生成记录，状态正确
- [ ] 埋点事件正常落库（scene_clicked、generation_saved、generation_shared 等）
- [ ] 后台指标接口 `/api/admin/metrics` 返回数据正常

## 本地联调与冒烟测试

```bash
# 1. 启动 PostgreSQL
docker compose -f infra/docker-compose.yml up -d

# 2. 运行迁移
psql $DATABASE_URL -f backend/migrations/0001_init.sql
psql $DATABASE_URL -f backend/migrations/0002_scene_templates.sql
psql $DATABASE_URL -f backend/migrations/0003_tracking_events.sql

# 3. 编译检查
cd backend && go build ./cmd/api && go build ./cmd/worker

# 4. 单元测试
go test ./...
```

**当前验证结果（2026-04-25）：**
- 后端编译：通过
- 单元测试：13 个包全部通过
- 容器联调：未执行（本地 Docker 未运行）

## 首发前文档确认

- [ ] 套餐价格表已确认（运营/商务）
- [ ] 5 个场景模板样例图已确认（设计）
- [ ] AI 生成结果审核文案已确认（法务/运营）
- [ ] 小程序隐私政策与用户协议已更新
- [ ] 微信支付商户号、API 证书已就绪
- [ ] COS 存储桶权限和 CDN 域名已配置

## 日常巡检

- 每日查看 `/api/admin/metrics` 中的支付趋势
- 监控 `generations` 表中 `failed` 状态比例
- 检查 `tracking_events` 增长量是否正常
