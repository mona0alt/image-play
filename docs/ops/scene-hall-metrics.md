# Scene Hall 运营指标定义

## Dashboard 接口

`GET /api/admin/metrics` 返回以下聚合指标：

```json
{
  "scene_clicks": {
    "portrait": 120,
    "invitation": 85
  },
  "generation_success": {
    "portrait": 45,
    "invitation": 30
  },
  "payments": {
    "2026-04-20": 12,
    "2026-04-21": 18,
    "2026-04-22": 15
  }
}
```

## 指标说明

### scene_clicks

- **定义**：用户点击场景入口的次数
- **数据来源**：`tracking_events` 表中 `event = 'scene_clicked'` 的记录
- **聚合维度**：按 `payload->>'scene_key'` 分组计数
- **用途**：衡量各场景的流量热度

### generation_success

- **定义**：图像生成成功的次数
- **数据来源**：`generations` 表中 `status = 'completed'` 的记录
- **聚合维度**：按 `scene_key` 分组计数
- **用途**：衡量各场景的实际转化效果

### payments

- **定义**：每日成功支付订单数
- **数据来源**：`orders` 表中 `status = 'paid'` 的记录
- **聚合维度**：按 `DATE(created_at)` 分组计数
- **用途**：衡量每日营收转化趋势

## 关键计算公式

| 指标 | 计算方式 |
|------|----------|
| 场景转化率 | `generation_success[scene] / scene_clicks[scene]` |
| 整体付费率 | `SUM(payments) / SUM(scene_clicks)` |
| 单场景 ARPU | `SUM(order_amount) / COUNT(distinct user_id)` |

## 告警建议

- `generation_success` 连续 1 小时为 0：检查 Worker 和 AI 服务
- `payments` 单日环比下降超过 50%：检查支付通道
- `scene_clicks` 某场景突然为 0：检查前端埋点和模板状态
