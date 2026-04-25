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

## 日常巡检

- 每日查看 `/api/admin/metrics` 中的支付趋势
- 监控 `generations` 表中 `failed` 状态比例
- 检查 `tracking_events` 增长量是否正常
