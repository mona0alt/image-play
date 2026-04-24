# AI 图片生成微信小程序 - 设计文档（V2）

## 1. 文档定位

### 1.1 文档目标
本文档用于指导 AI 图片生成微信小程序的 MVP 设计与实现，重点解决四类问题：

- 产品是否能形成清晰、可信的付费闭环
- 技术架构是否能在单机 MVP 条件下稳定运行
- 审核、支付、账务是否具备上线级别的一致性
- 后续是否能在不推翻架构的前提下继续扩展

### 1.2 设计结论
本项目采用 `单机 MVP，但按生产链路设计` 的方案：

- 前端保持轻量，强调低学习成本和低决策负担
- 后端采用 `API + Worker` 的异步任务模式
- 队列能力先落在 PostgreSQL，不在 MVP 引入 Redis / MQ
- 图片上传走 COS 直传，但必须有后端签发、回执确认和对象校验
- 模型调用前做文本审核，模型返回后做结果审核
- 计费采用“价格快照 + 幂等提交 + 事务扣费”的保守模型

### 1.3 MVP 成功标准

- 用户能在 3 分钟内完成首次登录、首次生成、首次保存
- 用户能清楚知道本次是否收费、为何失败、下一步怎么做
- 单个用户不会出现重复生成、重复扣费、扣费后查不到图
- 管理员能看到收入、生成成功率、审核拒绝率和异常任务

### 1.4 非目标
以下能力不纳入第一期：

- 社区、作品广场、点赞评论
- 多图批量生成、局部重绘、局部擦除
- PC Web 用户端
- 多模型自由切换
- 复杂会员体系、订阅制、优惠券

---

## 2. 产品定义

### 2.1 用户与场景
目标用户以轻量创作者和普通兴趣用户为主，典型场景包括：

- 快速生成头像、壁纸、配图
- 将自拍或参考图改造成特定风格
- 为朋友圈、小红书、社群内容生成图片素材

MVP 不追求专业设计工作流，而是追求：

- 上手快
- 成图快
- 失败可解释
- 支付可信

### 2.2 产品体验原则

1. **先让用户明白，再让用户点击**
   生成前就展示模式、费用、排队与审核说明，不把关键规则藏起来。
2. **先异步稳定，再追求“假同步”体验**
   用户感知可以很快，但服务端必须以任务状态为准，避免超时和重复提交。
3. **对新用户友好**
   首次使用不要求理解复杂参数，只给必要选项。
4. **对付费用户透明**
   每一笔消费都能解释到“为什么扣、扣了多少、对应哪次生成”。
5. **违规内容尽量前置拦截**
   降低模型成本，也减少“生成完才告诉用户不行”的挫败感。

---

## 3. MVP 功能范围

### 3.1 用户侧能力

- 微信登录
- 文生图
- 图生图
- 风格预设
- 异步生成任务与状态轮询
- 历史记录查看、搜索、软删除
- 结果图保存到相册、生成分享卡片
- 充值与余额管理
- 免费额度
- 隐私政策、用户协议确认
- 问题反馈入口

### 3.2 管理侧能力

- 管理员登录
- Dashboard 概览
- 用户列表与封禁查询
- 充值订单与流水查询
- 生成任务查询
- 审核拒绝任务查询
- 配置价格、免费额度、套餐、上传限制、风格开关
- 人工补单、人工调账、人工解封
- 操作审计日志

### 3.3 不做但预留接口的能力

- 重新生成沿用旧 prompt
- 异常任务重试
- 客服补偿额度
- AB 测试不同价格策略

---

## 4. 核心产品方案

### 4.1 首页结构
首页保持单页主工作台，不再拆多个页面来增加学习成本。

#### 顶部区

- 左上角：历史记录入口
- 中间：模式切换，`文生图 / 图生图`
- 右上角：个人入口，展示头像、余额、免费额度

#### 主输入区

- Prompt 输入框
- 风格选择器
- 图生图模式下显示“参考图卡片”
- 可选的负向提示词入口默认折叠

#### 底部操作区

- 主按钮：`立即生成`
- 副信息：`本次预计扣费`
- 任务状态提示：`排队中 / 生成中 / 审核中`

#### 体验细节

- 未同意协议时，生成按钮不可点击，并给出明确原因
- 免费额度可用时，费用显示为 `本次免费`
- 余额不足时，按钮文案改为 `充值后生成`
- 提交后按钮进入不可重复点击状态，避免用户连点

### 4.2 首次使用流程

1. 用户进入首页
2. 触发微信登录
3. 首次登录强制确认《隐私政策》《用户协议》
4. 展示新用户免费额度和一句话规则
5. 用户输入 prompt 后直接生成

首屏必须同时给出三条解释：

- 生成可能需要 10 到 40 秒
- 违规内容会被拦截
- 成功生成后才会消费免费额度或余额

### 4.3 图生图流程体验

1. 用户点击上传参考图
2. 先拿上传凭证，再直传 COS
3. 上传成功后先做原图审核
4. 审核通过后，参考图卡片状态改为 `可用`
5. 用户再输入 prompt 发起生成

如果原图审核拒绝：

- 不进入生成流程
- 不消耗免费额度
- 给出拒绝原因分类
- 引导用户重新上传

### 4.4 结果页体验
结果采用底部上拉的全屏卡片，而不是单独页面跳转，减少打断感。

结果区包含：

- 大图预览
- Prompt 摘要
- 保存到相册
- 分享卡片
- 再来一张
- 修改提示词

失败态不能只显示“生成失败”，必须区分：

- 文本不合规
- 原图不合规
- 模型超时
- 平台繁忙
- 系统异常

每种失败都要给一个下一步动作。

### 4.5 历史记录体验
历史记录不是简单列表，而是“任务与结果中心”。

列表项展示：

- 缩略图或失败图标
- prompt 前 20 个字
- 生成模式
- 状态
- 时间
- 消费信息

支持：

- 关键词搜索
- 按状态筛选
- 软删除
- 从历史记录重新发起

软删除语义：

- 对用户侧隐藏
- 不立刻物理删除底层对象
- 后台仍可审计

### 4.6 充值体验
充值页必须解决“为什么要充”和“充了能做什么”。

展示内容：

- 当前余额
- 剩余免费额度
- 套餐说明
- 单价说明
- 常见问题

推荐套餐不超过 3 个，避免选择过载。

### 4.7 审核与惩罚策略
MVP 采用“先拦截，后惩罚”的策略，而不是靠收费惩罚用户。

规则如下：

- `prompt` 审核不通过：不生成，不收费
- 原图审核不通过：不生成，不收费
- 模型失败或超时：不收费
- 结果图审核不通过：默认不收费
- 同一用户连续多次触发违规：短时封禁

封禁策略建议：

- 连续 3 次审核拒绝，封禁 10 分钟
- 连续 5 次审核拒绝，封禁 24 小时并进入人工复核名单

说明：
将“结果拒审仍收费”作为未来灰度策略，不作为 MVP 默认规则。MVP 以提升信任和转化为优先。

---

## 5. 业务规则

### 5.1 价格规则

- 基础价格：默认 1 元 / 张
- 价格按系统配置管理
- 每次创建任务时快照单价到任务记录
- 后续修改价格不影响已创建任务

### 5.2 免费额度规则

- 新用户默认 3 次免费额度
- 扣费顺序：先免费额度，再余额
- 免费额度只在任务成功完成时扣减
- 审核拒绝、模型失败、系统失败均不扣减

### 5.3 计费规则

- 成功生成且结果审核通过，才计为成功消费
- 单个生成任务最多产生一条正式消费流水
- 任意异常重试都不能导致重复扣费

### 5.4 并发规则

- 同一用户同时只允许 1 个活跃生成任务
- 活跃状态定义为：`queued`、`running`、`result_auditing`
- 同一用户重复点击提交，若 `client_request_id` 相同，则返回同一个任务

### 5.5 删除与留存规则

- 用户侧删除为软删除
- 默认保留任务记录长期可查
- 底层原图和结果图建议做冷热分层
- 软删除对象建议 180 天后归档，审计副本单独保留

---

## 6. 总体技术架构

### 6.1 架构图

```text
┌──────────────────────────────────────────────┐
│              微信小程序（UniApp）             │
└─────────────────────┬────────────────────────┘
                      │ HTTPS
┌─────────────────────▼────────────────────────┐
│                  Higress 网关                 │
│  /api/*    -> Go API                          │
│  /admin/*  -> 管理后台静态资源                │
│  /model/*  -> 外部模型 API                    │
└───────────────┬───────────────────┬──────────┘
                │                   │
     Docker 内网│                   │公网受控访问
┌───────────────▼───────┐   ┌──────▼────────────┐
│       Go API 服务      │   │    Go Worker      │
│  登录/支付/任务创建     │   │  领取任务/调模型   │
│  查询/管理后台 API      │   │  审核/落库/扣费    │
└───────────────┬────────┘   └────────┬─────────┘
                │                     │
        ┌───────▼────────┐   ┌────────▼─────────┐
        │ PostgreSQL 15   │   │ 腾讯云 COS / CI   │
        │ 业务库 + 任务队列 │   │ 对象存储 + 图片审核│
        └─────────────────┘   └──────────────────┘
```

### 6.2 组件职责

#### 小程序前端

- 承担输入、展示、支付调起、状态轮询
- 不直接接触模型 API
- 不直接信任本地状态判断是否成功

#### Higress

- 统一公网入口
- 统一模型出口
- 提供限流、鉴权、审计、日志埋点

#### Go API

- 登录鉴权
- 上传凭证签发
- 任务创建与查询
- 支付下单与回调
- 管理后台 API

#### Go Worker

- 从 PostgreSQL 领取待执行任务
- 调用模型
- 存储结果图
- 调用审核服务
- 在数据库事务中完成任务终态与扣费

#### PostgreSQL

- 业务主库
- 任务队列承载
- 账务流水
- 审计数据

### 6.3 为什么 MVP 不引入 Redis / MQ
本项目第一阶段目标是低成本上线，而不是高并发平台化。

采用 PostgreSQL 队列足够支撑：

- 单机部署
- 单用户限 1 并发
- 秒级轮询任务
- 低到中等请求量

但设计上保留后续升级路径：

- 任务领取接口与执行器解耦
- 未来可替换为 Redis Stream / Kafka / RabbitMQ

---

## 7. 任务处理架构

### 7.1 统一任务模型
文生图和图生图统一为 `generation` 任务，只在输入源上有差异。

统一字段包括：

- 用户信息
- 输入参数
- 价格快照
- 账务状态
- 执行状态
- 审核结果
- 结果对象引用

### 7.2 任务状态机

| 状态 | 说明 | 是否用户可见 |
|------|------|--------------|
| queued | 已创建，等待 worker 处理 | 是 |
| running | 已开始调用模型 | 是 |
| result_auditing | 已生成，等待结果审核 | 是 |
| succeeded | 成功完成 | 是 |
| rejected | 因审核拒绝结束 | 是 |
| failed | 因模型/系统失败结束 | 是 |
| cancelled | 人工取消或系统取消 | 是 |

约束：

- 终态只有 `succeeded`、`rejected`、`failed`、`cancelled`
- 终态一旦写入，不允许再回到处理中状态
- 用户轮询只认任务状态，不认 HTTP 请求是否超时

### 7.3 图生图资产状态机
上传资产单独建模，不把上传状态塞进生成任务。

| 状态 | 说明 |
|------|------|
| pending_upload | 已申请上传，用户尚未完成直传 |
| uploaded | COS 已收到对象，待后端确认 |
| source_auditing | 原图审核中 |
| approved | 原图可用于生成 |
| rejected | 原图审核拒绝 |
| expired | 上传超时未完成 |

### 7.4 文生图流程

1. 客户端提交 `POST /api/generations`
2. API 校验协议、封禁、额度、幂等键
3. API 在事务中创建 `queued` 任务
4. Worker 领取任务并改为 `running`
5. 通过 Higress 调用模型
6. 结果图入 COS
7. 状态进入 `result_auditing`
8. 审核通过后事务性扣费并置为 `succeeded`
9. 审核拒绝则置为 `rejected`

### 7.5 图生图流程

1. 客户端申请上传凭证
2. 客户端直传 COS
3. 客户端提交上传完成回执
4. API 校验对象元数据并触发原图审核
5. 原图审核通过后，用户才能创建生成任务
6. 后续流程与文生图一致

### 7.6 轮询策略

- 提交任务后立即返回 `generation_id`
- 前端前 15 秒每 2 秒轮询一次
- 15 秒后每 3 到 5 秒轮询一次
- 到达终态后停止轮询

前端不做“超时即失败”的本地裁决，只提示：

- `生成时间较长，正在继续处理`

---

## 8. 计费与账务设计

### 8.1 计费原则

- 价格在任务创建时快照
- 成功才消费
- 失败与审核拒绝不消费
- 消费必须能追溯到具体任务
- 支付与消费都必须幂等

### 8.2 为什么不采用“先扣费，再失败退款”
MVP 更适合“成功后扣费”的模型，原因是：

- 更符合用户直觉
- 能降低投诉和客服成本
- 任务并发受限为 1，账务风险可控

但要满足两个前提：

- 只能有 1 个活跃任务
- 扣费必须在数据库事务内完成

### 8.3 账务模型
`users.balance` 和 `users.free_quota` 作为汇总字段保留，`transactions` 是对账权威来源。

流水类型建议如下：

- `recharge`
- `bonus`
- `consume`
- `refund`
- `manual_adjust`

消费流水必须满足：

- `generation_id` 唯一绑定
- 同一 `generation_id` 最多 1 条 `consume`

### 8.4 任务内计费字段
在 `generations` 表中保留以下字段：

- `unit_price`
- `charge_mode`，取值：`free_quota` / `balance`
- `billing_status`，取值：`pending` / `charged` / `released`

说明：

- 任务创建成功时，`billing_status = pending`
- 任务成功完成时，写消费流水并改为 `charged`
- 任务失败或拒绝时，改为 `released`

### 8.5 扣费事务
任务成功后，worker 必须在一个数据库事务内完成以下操作：

1. 锁定用户行
2. 校验该任务尚未扣费
3. 扣减免费额度或余额
4. 写入消费流水
5. 将任务改为 `succeeded + charged`

只要事务未提交，任务就不能对用户显示为最终成功。

### 8.6 支付回调幂等
微信支付回调需满足：

- 基于 `wx_order_no` 唯一
- 金额校验
- 重复回调只处理一次
- 处理成功后写 `recharge` 与 `bonus` 流水

---

## 9. 数据模型设计

### 9.1 users

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | 用户 ID |
| openid | VARCHAR(64) UNIQUE | 微信 openid |
| unionid | VARCHAR(64) NULL | 微信 unionid |
| nickname | VARCHAR(64) NULL | 昵称 |
| avatar_url | VARCHAR(255) NULL | 头像 |
| balance | DECIMAL(10,2) DEFAULT 0.00 | 当前余额 |
| free_quota | INT DEFAULT 3 | 剩余免费额度 |
| audit_reject_streak | INT DEFAULT 0 | 连续审核拒绝次数 |
| ban_until | TIMESTAMP NULL | 封禁截止时间 |
| accepted_policy_at | TIMESTAMP NULL | 同意协议时间 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

### 9.2 assets

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | 资源 ID |
| user_id | BIGINT FK | 用户 ID |
| kind | VARCHAR(16) | source / result |
| object_key | VARCHAR(255) UNIQUE | COS 对象 key |
| mime_type | VARCHAR(64) | 文件类型 |
| size_bytes | BIGINT | 文件大小 |
| width | INT NULL | 宽 |
| height | INT NULL | 高 |
| sha256 | VARCHAR(64) NULL | 内容摘要 |
| status | VARCHAR(20) | 见资产状态机 |
| audit_label | VARCHAR(64) NULL | 审核标签 |
| audit_reason | VARCHAR(255) NULL | 审核原因 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |
| deleted_at | TIMESTAMP NULL | 软删除 |

索引建议：

- `user_id + kind + created_at`
- `status + created_at`

### 9.3 generations

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | 任务 ID |
| user_id | BIGINT FK | 用户 ID |
| client_request_id | VARCHAR(64) | 客户端幂等键 |
| type | VARCHAR(16) | text2img / img2img |
| prompt | TEXT | 用户提示词 |
| negative_prompt | TEXT NULL | 负向提示词 |
| style_id | VARCHAR(32) | 风格 ID |
| source_asset_id | BIGINT NULL FK | 原图资产 |
| result_asset_id | BIGINT NULL FK | 结果资产 |
| status | VARCHAR(20) | 见任务状态机 |
| failure_code | VARCHAR(64) NULL | 失败码 |
| failure_message | VARCHAR(255) NULL | 用户可见失败信息 |
| unit_price | DECIMAL(10,2) | 价格快照 |
| charge_mode | VARCHAR(16) | free_quota / balance |
| billing_status | VARCHAR(16) | pending / charged / released |
| model_name | VARCHAR(64) | 模型名称 |
| provider_request_id | VARCHAR(128) NULL | 模型侧请求 ID |
| queued_at | TIMESTAMP | 入队时间 |
| started_at | TIMESTAMP NULL | 开始时间 |
| completed_at | TIMESTAMP NULL | 完成时间 |
| latency_ms | INT NULL | 总耗时 |
| deleted_at | TIMESTAMP NULL | 软删除 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

关键约束建议：

- 唯一索引：`user_id + client_request_id`
- 部分唯一索引：同一 `user_id` 在活跃状态下只能有 1 条任务
- 唯一索引：`result_asset_id` 可空但唯一

### 9.4 orders

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | |
| user_id | BIGINT FK | |
| order_no | VARCHAR(64) UNIQUE | 平台订单号 |
| wx_order_no | VARCHAR(64) UNIQUE NULL | 微信订单号 |
| package_code | VARCHAR(32) | 套餐编码 |
| amount_payable | DECIMAL(10,2) | 用户实付 |
| amount_credit | DECIMAL(10,2) | 到账金额 |
| bonus_credit | DECIMAL(10,2) | 赠送金额 |
| status | VARCHAR(16) | created / paid / failed / closed |
| paid_at | TIMESTAMP NULL | |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

### 9.5 transactions

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | |
| user_id | BIGINT FK | |
| order_id | BIGINT NULL FK | 关联订单 |
| generation_id | BIGINT NULL FK | 关联任务 |
| type | VARCHAR(20) | recharge / bonus / consume / refund / manual_adjust |
| amount | DECIMAL(10,2) | 正数充值，负数消费 |
| balance_after | DECIMAL(10,2) | 操作后余额 |
| free_quota_after | INT | 操作后免费额度 |
| remark | VARCHAR(255) NULL | 备注 |
| created_at | TIMESTAMP | |

关键约束建议：

- `generation_id + type='consume'` 唯一
- `order_id + type in ('recharge','bonus')` 可用于幂等核对

### 9.6 system_configs

| 字段 | 类型 | 说明 |
|------|------|------|
| config_key | VARCHAR(64) PK | 配置键 |
| config_value | TEXT | 配置值，JSON 或字符串 |
| updated_at | TIMESTAMP | |
| updated_by | BIGINT NULL | 管理员 ID |

### 9.7 admin_users

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | |
| username | VARCHAR(32) UNIQUE | |
| password_hash | VARCHAR(255) | bcrypt |
| role | VARCHAR(16) | viewer / operator / admin / super_admin |
| last_login_at | TIMESTAMP NULL | |
| created_at | TIMESTAMP | |

### 9.8 admin_audit_logs

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | |
| admin_user_id | BIGINT FK | 管理员 |
| action | VARCHAR(64) | 登录、改价、补单、调账、解封等 |
| target_type | VARCHAR(32) | user / order / generation / config |
| target_id | BIGINT NULL | |
| detail | JSONB | 变更详情 |
| created_at | TIMESTAMP | |

---

## 10. API 设计

### 10.1 认证

- 小程序端：JWT Bearer Token
- 管理后台：独立 admin token，短有效期 + 刷新机制

### 10.2 用户端 API

#### 认证与用户

- `POST /api/auth/login`：微信登录
- `POST /api/policies/accept`：确认隐私政策与用户协议
- `GET /api/me`：当前用户信息

#### 资产上传

- `POST /api/assets/upload-intents`：申请上传凭证
- `POST /api/assets/{id}/complete`：上传完成回执
- `GET /api/assets/{id}`：查询资产状态

#### 生成任务

- `POST /api/generations`：创建生成任务
- `GET /api/generations/{id}`：查询任务详情
- `GET /api/generations`：历史任务列表
- `POST /api/generations/{id}/regenerate`：基于历史任务重新创建
- `DELETE /api/generations/{id}`：软删除历史任务

#### 支付

- `GET /api/pay/packages`：获取充值套餐
- `POST /api/pay/orders`：创建充值订单
- `POST /api/pay/callback`：微信支付回调

#### 辅助能力

- `GET /api/configs/client`：获取前端必要配置
- `POST /api/feedback`：问题反馈

### 10.3 管理端 API

- `POST /api/admin/login`
- `GET /api/admin/dashboard`
- `GET /api/admin/users`
- `GET /api/admin/orders`
- `GET /api/admin/transactions`
- `GET /api/admin/generations`
- `GET /api/admin/assets`
- `GET /api/admin/configs`
- `PUT /api/admin/configs`
- `POST /api/admin/users/{id}/unban`
- `POST /api/admin/orders/{id}/repair`
- `POST /api/admin/users/{id}/adjust-balance`
- `GET /api/admin/audit-logs`

### 10.4 关键接口约束

#### 创建任务请求

```json
{
  "client_request_id": "b60f2c53-cc80-4c4d-8a8b-f53d4c3a0e10",
  "type": "img2img",
  "prompt": "把这张人物照片改成吉卜力动画风格",
  "negative_prompt": "模糊, 低清晰度",
  "style_id": "ghibli",
  "source_asset_id": 12345
}
```

返回示例：

```json
{
  "code": "OK",
  "message": "accepted",
  "data": {
    "generation_id": 98765,
    "status": "queued",
    "estimated_price": "1.00",
    "charge_mode": "free_quota"
  }
}
```

#### 查询任务返回

```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "id": 98765,
    "status": "result_auditing",
    "failure_code": null,
    "failure_message": null,
    "result_image_url": null,
    "charge_mode": "free_quota",
    "billing_status": "pending"
  }
}
```

### 10.5 错误码
建议统一错误模型：

```json
{
  "code": "INSUFFICIENT_BALANCE",
  "message": "余额不足，请先充值",
  "data": null,
  "request_id": "req_20260424_xxx"
}
```

常见错误码：

- `UNAUTHORIZED`
- `FORBIDDEN`
- `POLICY_NOT_ACCEPTED`
- `BANNED`
- `INVALID_PARAM`
- `SOURCE_ASSET_NOT_READY`
- `ACTIVE_GENERATION_EXISTS`
- `INSUFFICIENT_BALANCE`
- `PROMPT_REJECTED`
- `SOURCE_IMAGE_REJECTED`
- `RESULT_REJECTED`
- `MODEL_TIMEOUT`
- `SYSTEM_BUSY`

---

## 11. 一致性、幂等与并发控制

### 11.1 幂等策略

- 所有创建型接口支持幂等键
- 生成任务使用 `client_request_id`
- 支付订单使用 `order_no`
- 微信回调使用 `wx_order_no`

### 11.2 用户并发控制
不通过“先查后写”判断活跃任务，而是通过数据库约束确保：

- 同一用户同一时刻最多 1 个活跃任务

推荐 PostgreSQL 部分唯一索引思路：

- `unique(user_id) where status in ('queued','running','result_auditing')`

### 11.3 Worker 领取任务
Worker 通过如下思路领取任务：

- 事务中查询 `queued` 任务
- `FOR UPDATE SKIP LOCKED`
- 成功领取后改为 `running`

这样可以保证未来即使扩容多个 worker，也不会重复执行同一任务。

### 11.4 外部调用失败处理

- 模型调用失败：有限次重试
- COS 上传回执校验失败：标记资产异常
- 审核服务超时：允许短重试，超过阈值后置为系统失败

重试原则：

- 不重试用户输入错误
- 谨慎重试第三方短时异常
- 所有重试都不得重复扣费

---

## 12. 安全与合规

### 12.1 内容审核链路
审核必须覆盖四类对象：

- 用户 `prompt`
- 用户上传原图
- 模型生成结果图
- 分享卡片文案

处理顺序：

1. `prompt` 审核
2. 原图审核
3. 模型生成
4. 结果图审核

### 12.2 对象存储安全

- COS 使用私有桶
- 结果图访问使用短时签名 URL
- 上传签名必须绑定用户、对象 key、大小、MIME、过期时间
- 上传完成后必须由后端二次校验对象元数据

### 12.3 支付安全

- 微信回调验签
- 金额校验
- 订单状态幂等
- 支付日志留存
- 异常补单必须有审计日志

### 12.4 管理后台安全

- RBAC 角色分级
- 登录限流
- 强密码策略
- 敏感操作二次确认
- 所有高风险操作写入审计日志

### 12.5 合规文档

- 首次登录必须确认《隐私政策》《用户协议》
- 后台提供投诉与处理链路
- 用户删除记录应有明确留存说明

---

## 13. 运维与可观测性

### 13.1 部署方案
单机部署仍然保留独立进程角色：

- `app-api`
- `app-worker`
- `higress`
- `postgres`

避免将“API 服务”和“任务执行”完全耦合在同一个请求生命周期里。

### 13.2 最低可观测性要求

- API 请求日志
- 支付回调日志
- 任务状态流转日志
- 模型调用日志
- 审核调用日志
- 错误告警

### 13.3 关键指标

- 登录成功率
- 首次生成成功率
- 平均生成耗时
- 任务终态分布
- 审核拒绝率
- 支付成功率
- 重复提交命中率
- 每日收入

### 13.4 告警建议

- 支付回调失败数异常
- 连续 5 分钟生成成功率过低
- 队列积压过多
- 数据库连接池耗尽
- 磁盘使用率过高

### 13.5 备份与恢复

- PostgreSQL 每日备份到 COS
- 每周做一次恢复演练
- 记录 RPO / RTO 目标

建议目标：

- RPO：24 小时以内
- RTO：4 小时以内

---

## 14. 管理后台设计

### 14.1 Dashboard 指标

- 今日收入
- 今日支付订单数
- 今日生成任务数
- 成功率
- 平均耗时
- 新增用户数
- 审核拒绝率
- 当前活跃任务数

### 14.2 关键运营视图

- 用户视图：余额、免费额度、封禁状态、最近任务
- 任务视图：状态、失败原因、耗时、审核结果
- 订单视图：支付状态、金额、回调状态
- 流水视图：充值、赠送、消费、调账

### 14.3 高风险操作
以下操作必须记录审计日志：

- 修改价格
- 修改套餐
- 人工补单
- 人工调账
- 解封用户
- 导出数据

---

## 15. 实施建议

### 15.1 必须先打通的主链路

- 登录、协议确认、用户信息初始化
- `prompt` 审核、任务创建、轮询查询
- worker 领取任务、调用模型、结果审核
- 成功扣费、失败不扣费、历史可查询
- 充值下单、支付回调、余额到账

说明：
这部分完成前，不建议投入大量时间做视觉打磨，因为产品是否成立，首先取决于主链路是否稳定。

### 15.2 第二阶段

- 增加图生图
- 增加搜索、筛选、重新生成
- 增加封禁与人工解封
- 增加审计日志

### 15.3 第三阶段

- 增加任务告警
- 增加客服补偿能力
- 灰度更复杂价格策略
- 增加更细粒度运营报表

---

## 16. 最终结论
本设计将原方案从“同步接口 + 结果后处理”的思路，升级为“异步任务 + 事务扣费 + 审核前置 + 可观测运维”的上线型 MVP 方案。

该方案的核心价值在于：

- 产品上更可信，用户更容易理解是否收费与为何失败
- 架构上更稳，避免超时、重复生成、重复扣费
- 运维上更可控，便于排障、补单、审计与扩容

如果后续进入实施阶段，建议先按本文档拆出两条主线：

- 用户生成主链路
- 支付与账务主链路

这两条链路优先级最高，且必须优先通过联调与异常测试。
