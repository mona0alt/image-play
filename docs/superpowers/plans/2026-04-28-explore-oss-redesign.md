# 发现页 OSS 运营内容改造 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将发现页数据源从用户生成的 `generations` 改为 OSS `explore/` 目录下的运营图片，保留全屏滑动浏览、点赞和拍同款交互。

**Architecture:** 数据库新增 `explore_assets`（运营内容）和 `explore_likes`（点赞）两张表；后端 `ExploreFeedHandler` / `ExploreLikeHandler` 改为查询新表；前端 `api.ts` 调整点赞接口参数；提供一次性 Go 脚本扫描 OSS `explore/` 前缀并入库。

**Tech Stack:** Go, Gin, PostgreSQL, uni-app (Vue 3), 阿里云 OSS Go SDK

---

## 文件映射

| 文件 | 操作 | 职责 |
|---|---|---|
| `backend/internal/migration/migration.go` | 修改 | 添加 `explore_assets`、`explore_likes` 表创建语句 |
| `backend/internal/config/config.go` | 修改 | `Config` 结构体补充 OSS 字段，支持环境变量覆盖 |
| `backend/internal/http/handlers/explore_handler.go` | 修改 | `ExploreFeedHandler`、`ExploreLikeHandler` 改为查询新表 |
| `frontend/src/services/api.ts` | 修改 | `likeExploreItem` 参数从 `generation_id` 改为 `explore_asset_id` |
| `frontend/src/components/explore/ExploreCard.vue` | 修改 | 隐藏作者信息区（运营内容无作者），保留点赞和拍同款按钮 |
| `backend/scripts/sync_explore_assets.go` | 创建 | 临时脚本：扫描 OSS `explore/` 目录，将图片 URL 写入 `explore_assets` |
| `backend/go.mod` | 修改 | 添加 `github.com/aliyun/aliyun-oss-go-sdk v2.2.9+incompatible` 依赖 |

---

### Task 1: 数据库迁移 — 新增运营内容表和点赞表

**Files:**
- Modify: `backend/internal/migration/migration.go:103-119`

- [ ] **Step 1: 在 migration.go 的 migrations 切片末尾追加新表定义**

```go
	`CREATE TABLE IF NOT EXISTS explore_assets (
		id BIGSERIAL PRIMARY KEY,
		image_url VARCHAR(500) NOT NULL,
		scene_key VARCHAR(32) NOT NULL DEFAULT '',
		prompt TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_explore_assets_image_url ON explore_assets(image_url);`,
	`CREATE TABLE IF NOT EXISTS explore_likes (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id),
		explore_asset_id BIGINT NOT NULL REFERENCES explore_assets(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (user_id, explore_asset_id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_explore_likes_asset_id ON explore_likes(explore_asset_id);`,
	`CREATE INDEX IF NOT EXISTS idx_explore_likes_user_id ON explore_likes(user_id);`,
```

- [ ] **Step 2: 验证编译通过**

Run: `cd /root/project/image-play/backend && go build ./...`
Expected: 无错误

- [ ] **Step 3: 运行迁移验证表创建**

Run: `cd /root/project/image-play && make dev-db && ./start-backend.sh api`
Expected: 服务正常启动，日志显示 migrations applied

- [ ] **Step 4: Commit**

```bash
git add backend/internal/migration/migration.go
git commit -m "feat(explore): add explore_assets and explore_likes tables"
```

---

### Task 2: 后端配置 — 补充 OSS 字段

**Files:**
- Modify: `backend/internal/config/config.go`

- [ ] **Step 1: 修改 Config 结构体，增加 OSS 字段**

```go
type Config struct {
	AppEnv          string `yaml:"app_env"`
	Port            string `yaml:"port"`
	DatabaseURL     string `yaml:"database_url"`
	JWTSecret       string `yaml:"jwt_secret"`
	WechatAppID     string `yaml:"wechat_app_id"`
	WechatAppSecret string `yaml:"wechat_app_secret"`
	OSSEndpoint     string `yaml:"oss_endpoint"`
	OSSBucket       string `yaml:"oss_bucket"`
	OSSAccessKeyID  string `yaml:"oss_access_key_id"`
	OSSAccessKeySecret string `yaml:"oss_access_key_secret"`
}
```

- [ ] **Step 2: 在 Load() 中添加 OSS 字段的环境变量覆盖**

在 `loadFromEnv()` 和 `Load()` 的 env override 段中补充：

```go
	if v := os.Getenv("OSS_ENDPOINT"); v != "" {
		cfg.OSSEndpoint = v
	}
	if v := os.Getenv("OSS_BUCKET"); v != "" {
		cfg.OSSBucket = v
	}
	if v := os.Getenv("OSS_ACCESS_KEY_ID"); v != "" {
		cfg.OSSAccessKeyID = v
	}
	if v := os.Getenv("OSS_ACCESS_KEY_SECRET"); v != "" {
		cfg.OSSAccessKeySecret = v
	}
```

并在 `loadFromEnv()` 中补充默认值：

```go
		OSSEndpoint:        getEnv("OSS_ENDPOINT", ""),
		OSSBucket:          getEnv("OSS_BUCKET", ""),
		OSSAccessKeyID:     getEnv("OSS_ACCESS_KEY_ID", ""),
		OSSAccessKeySecret: getEnv("OSS_ACCESS_KEY_SECRET", ""),
```

- [ ] **Step 3: 验证编译**

Run: `cd /root/project/image-play/backend && go build ./...`
Expected: 无错误

- [ ] **Step 4: Commit**

```bash
git add backend/internal/config/config.go
git commit -m "feat(config): add OSS fields to Config struct"
```

---

### Task 3: 后端 Handler — 改造发现页查询和点赞逻辑

**Files:**
- Modify: `backend/internal/http/handlers/explore_handler.go`

- [ ] **Step 1: 替换 ExploreFeedHandler 的查询逻辑**

将 `ExploreFeedHandler` 中从 `generations` / `users` / `likes` 三表 JOIN 的查询，替换为对 `explore_assets` 的单表查询 + `explore_likes` 的聚合子查询。

替换 SQL 查询部分（约第 44-56 行）：

```go
		query := `
			SELECT ea.id, ea.image_url, ea.scene_key, ea.prompt, ea.created_at,
				COALESCE(l.cnt, 0) as like_count
			FROM explore_assets ea
			LEFT JOIN (
				SELECT explore_asset_id, COUNT(*) as cnt FROM explore_likes GROUP BY explore_asset_id
			) l ON l.explore_asset_id = ea.id
			ORDER BY ea.created_at DESC
			LIMIT $1 OFFSET $2
		`
```

替换 rows.Scan 部分（约第 66-84 行）：

```go
	for rows.Next() {
		var item ExploreItem
		var createdAt sql.NullTime
		err := rows.Scan(
			&item.ID, &item.ImageURL, &item.SceneKey, &item.Prompt, &createdAt,
			&item.LikeCount,
		)
		if err != nil {
			continue
		}
		item.ThumbnailURL = item.ImageURL
		item.Description = ""
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z")
		}

		// Check if current user liked this item
		if uid, exists := c.Get("user_id"); exists {
			if userIDVal, ok := uid.(int64); ok {
				var existsLike bool
				_ = db.QueryRowContext(c.Request.Context(),
					"SELECT EXISTS(SELECT 1 FROM explore_likes WHERE user_id = $1 AND explore_asset_id = $2)",
					userIDVal, item.ID,
				).Scan(&existsLike)
				item.IsLiked = existsLike
			}
		}

		items = append(items, item)
	}
```

替换 countQuery（约第 110 行）：

```go
	countQuery := `SELECT COUNT(*) FROM explore_assets`
```

注意：返回的 JSON 结构保持不变，但 `user` 字段现在没有实际意义，可以在返回时给一个占位值：

在 items append 前补充：

```go
		item.User = ExploreUser{
			ID:        "0",
			Nickname:  "",
			AvatarURL: "",
		}
```

- [ ] **Step 2: 替换 ExploreLikeHandler 的表名和字段名**

修改 `LikeRequest` 结构体（约第 125-128 行）：

```go
type LikeRequest struct {
	ExploreAssetID int64  `json:"explore_asset_id"`
	Action         string `json:"action"`
}
```

替换 like 插入 SQL（约第 154-158 行）：

```go
		_, err := db.ExecContext(c.Request.Context(),
			"INSERT INTO explore_likes (user_id, explore_asset_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			uid, req.ExploreAssetID,
		)
```

替换 unlike 删除 SQL（约第 164-167 行）：

```go
		_, err := db.ExecContext(c.Request.Context(),
			"DELETE FROM explore_likes WHERE user_id = $1 AND explore_asset_id = $2",
			uid, req.ExploreAssetID,
		)
```

替换 like count 查询（约第 175-177 行）：

```go
	_ = db.QueryRowContext(c.Request.Context(),
		"SELECT COUNT(*) FROM explore_likes WHERE explore_asset_id = $1", req.ExploreAssetID,
	).Scan(&count)
```

- [ ] **Step 3: 验证编译**

Run: `cd /root/project/image-play/backend && go build ./...`
Expected: 无错误

- [ ] **Step 4: Commit**

```bash
git add backend/internal/http/handlers/explore_handler.go
git commit -m "feat(explore): switch feed and like handlers to explore_assets table"
```

---

### Task 4: 前端 API — 更新点赞接口参数

**Files:**
- Modify: `frontend/src/services/api.ts`

- [ ] **Step 1: 修改 likeExploreItem 函数签名和请求体**

将第 237-244 行改为：

```typescript
export function likeExploreItem(exploreAssetId: number, action: 'like' | 'unlike') {
  return request<{ success: boolean; like_count: number }>({
    url: '/api/explore/like',
    method: 'POST',
    data: { explore_asset_id: exploreAssetId, action },
    headers: { 'Content-Type': 'application/json' },
  })
}
```

- [ ] **Step 2: 验证类型检查**

Run: `cd /root/project/image-play/frontend && npm run type-check`
Expected: 无类型错误

- [ ] **Step 3: Commit**

```bash
git add frontend/src/services/api.ts
git commit -m "feat(api): update likeExploreItem to use explore_asset_id"
```

---

### Task 5: 前端卡片 — 隐藏作者信息区

**Files:**
- Modify: `frontend/src/components/explore/ExploreCard.vue`

- [ ] **Step 1: 隐藏 card__info 区域（运营内容无作者信息）**

将模板中 `card__info` 整个 view 用 `v-if="false"` 包裹或注释掉，保留 `card__actions`（点赞和拍同款按钮）：

```vue
    <!-- Bottom-left info card -->
    <view v-if="false" class="card__info">
      ...
    </view>
```

或直接从模板中删除 `card__info` 及其内部内容。

- [ ] **Step 2: 验证编译**

Run: `cd /root/project/image-play/frontend && npm run type-check`
Expected: 无类型错误

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/explore/ExploreCard.vue
git commit -m "feat(explore): hide author info in explore cards for curated content"
```

---

### Task 6: OSS 同步脚本

**Files:**
- Create: `backend/scripts/sync_explore_assets.go`

- [ ] **Step 1: 添加阿里云 OSS SDK 依赖**

Run:
```bash
cd /root/project/image-play/backend
go get github.com/aliyun/aliyun-oss-go-sdk/oss
```

- [ ] **Step 2: 创建同步脚本文件**

```go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"image-play/internal/config"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	if cfg.OSSBucket == "" || cfg.OSSEndpoint == "" || cfg.OSSAccessKeyID == "" || cfg.OSSAccessKeySecret == "" {
		log.Fatal("OSS config missing. Check config.yaml or env vars.")
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	client, err := oss.New(cfg.OSSEndpoint, cfg.OSSAccessKeyID, cfg.OSSAccessKeySecret)
	if err != nil {
		log.Fatalf("oss new: %v", err)
	}

	bucket, err := client.Bucket(cfg.OSSBucket)
	if err != nil {
		log.Fatalf("oss bucket: %v", err)
	}

	marker := ""
	prefix := "explore/"
	inserted := 0
	skipped := 0

	for {
		res, err := bucket.ListObjects(oss.Marker(marker), oss.Prefix(prefix), oss.MaxKeys(100))
		if err != nil {
			log.Fatalf("list objects: %v", err)
		}

		for _, obj := range res.Objects {
			if !strings.HasSuffix(strings.ToLower(obj.Key), ".jpg") && !strings.HasSuffix(strings.ToLower(obj.Key), ".jpeg") {
				continue
			}

			imageURL := fmt.Sprintf("https://%s.%s/%s", cfg.OSSBucket, cfg.OSSEndpoint, obj.Key)

			var exists int
			err := db.QueryRow("SELECT 1 FROM explore_assets WHERE image_url = $1", imageURL).Scan(&exists)
			if err == nil {
				skipped++
				continue
			}

			_, err = db.Exec(
				"INSERT INTO explore_assets (image_url, scene_key, prompt) VALUES ($1, $2, $3)",
				imageURL, "", "",
			)
			if err != nil {
				log.Printf("insert failed for %s: %v", imageURL, err)
				continue
			}
			inserted++
			log.Printf("inserted: %s", imageURL)
		}

		if !res.IsTruncated {
			break
		}
		marker = res.NextMarker
	}

	log.Printf("Done. Inserted: %d, Skipped: %d", inserted, skipped)
}
```

- [ ] **Step 3: 验证编译**

Run: `cd /root/project/image-play/backend && go build ./scripts/sync_explore_assets.go`
Expected: 无错误

- [ ] **Step 4: Commit**

```bash
git add backend/go.mod backend/go.sum backend/scripts/sync_explore_assets.go
git commit -m "feat(explore): add OSS sync script for explore_assets"
```

---

### Task 7: 端到端验证

- [ ] **Step 1: 确保数据库已启动**

Run: `cd /root/project/image-play && make dev-db`

- [ ] **Step 2: 启动后端**

Run: `./start-backend.sh api`
Expected: 服务正常启动在 :8080

- [ ] **Step 3: 运行同步脚本（确保 OSS 中有测试图片）**

Run:
```bash
cd /root/project/image-play/backend
go run ./scripts/sync_explore_assets.go
```
Expected: 输出 `Done. Inserted: N, Skipped: M`

- [ ] **Step 4: 验证数据库中有数据**

Run: `docker exec -it image-play-postgres psql -U postgres -d image_play -c "SELECT COUNT(*) FROM explore_assets;"`
Expected: 返回数量 > 0

- [ ] **Step 5: 测试发现页 API**

Run:
```bash
curl -H "Authorization: Bearer <valid-token>" \
  "http://localhost:8080/api/explore/feed?page=1&page_size=10"
```
Expected: 返回 JSON，items 数组中包含 `id`、`image_url`、`scene_key`、`prompt`、`like_count`、`is_liked`

- [ ] **Step 6: 测试点赞 API**

Run:
```bash
curl -H "Authorization: Bearer <valid-token>" \
  -H "Content-Type: application/json" \
  -X POST \
  -d '{"explore_asset_id": 1, "action": "like"}' \
  http://localhost:8080/api/explore/like
```
Expected: 返回 `{"success": true, "like_count": 1}`

- [ ] **Step 7: 前端构建验证**

Run: `cd /root/project/image-play && make dev-frontend`
Expected: 构建成功，无错误

- [ ] **Step 8: Commit 任何收尾改动**

---

## Self-Review Checklist

### Spec coverage
- [x] 新增 `explore_assets` 表 → Task 1
- [x] 新增 `explore_likes` 表 → Task 1
- [x] 后端 Handler 改造 → Task 3
- [x] 前端 API 调整 → Task 4
- [x] 前端卡片隐藏作者信息 → Task 5
- [x] 同步脚本 → Task 6
- [x] OSS 配置读取 → Task 2 + Task 6

### Placeholder scan
- [x] 无 "TBD"、"TODO"、"implement later"
- [x] 所有步骤包含实际代码
- [x] 无 "appropriate error handling" 等模糊描述

### Type consistency
- [x] `explore_asset_id` 在前后端命名一致
- [x] `LikeRequest` 的 JSON tag 与前端请求体一致
