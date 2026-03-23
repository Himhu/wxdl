# 兑换码使用情况查询 — API 对接文档

## 1. 认证方式

所有兑换码接口需要 **管理员权限**（Admin）。通过 `Authorization` 请求头传入 Access Token：

```
Authorization: <your_access_token>
```

Access Token 可在系统后台「个人设置」页面生成。

---

## 2. 接口列表

### 2.1 分页获取所有兑换码

```
GET /api/redemption/?p={page}&page_size={size}
```

**参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `p` | int | 1 | 页码 |
| `page_size` | int | 10 | 每页条数（最大100） |

**响应示例：**

```json
{
  "success": true,
  "data": {
    "page": 1,
    "page_size": 10,
    "total": 50,
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "key": "a1b2c3d4...",
        "status": 3,
        "name": "活动赠送",
        "quota": 500000,
        "created_time": 1710000000,
        "redeemed_time": 1710100000,
        "used_user_id": 42,
        "expired_time": 0
      }
    ]
  }
}
```

### 2.2 搜索兑换码

```
GET /api/redemption/search?keyword={keyword}&p={page}&page_size={size}
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `keyword` | string | 按名称前缀匹配，或精确匹配 ID |

响应结构同 2.1。

### 2.3 查询单个兑换码

```
GET /api/redemption/{id}
```

**响应示例：**

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 1,
    "user_id": 1,
    "key": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "status": 3,
    "name": "活动赠送",
    "quota": 500000,
    "created_time": 1710000000,
    "redeemed_time": 1710100000,
    "used_user_id": 42,
    "expired_time": 0
  }
}
```

---

## 3. 字段说明

### 3.1 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | 兑换码 ID |
| `user_id` | int | 创建者用户 ID |
| `key` | string | 兑换码密钥（32位 UUID） |
| `status` | int | 状态，见下表 |
| `name` | string | 兑换码名称 |
| `quota` | int | 额度（内部单位） |
| `created_time` | int64 | 创建时间（Unix 时间戳，秒） |
| `redeemed_time` | int64 | 兑换时间（Unix 时间戳，未使用时为 0） |
| `used_user_id` | int | 兑换人用户 ID（未使用时为 0） |
| `expired_time` | int64 | 过期时间（Unix 时间戳，0 表示永不过期） |

### 3.2 状态值

| status | 含义 |
|--------|------|
| 1 | 未使用 |
| 2 | 已禁用 |
| 3 | 已使用 |

> 额外判断：`status == 1` 且 `expired_time != 0` 且 `expired_time < 当前时间戳` → 已过期

---

## 4. 对接示例

### 4.1 Python — 遍历所有兑换码

```python
import requests
import time

BASE_URL = "https://your-domain.com"
HEADERS = {"Authorization": "your_access_token"}


def get_all_redemptions():
    page, all_items = 1, []
    while True:
        resp = requests.get(
            f"{BASE_URL}/api/redemption/",
            params={"p": page, "page_size": 100},
            headers=HEADERS,
        ).json()
        items = resp["data"]["items"]
        all_items.extend(items)
        if len(all_items) >= resp["data"]["total"]:
            break
        page += 1
    return all_items
```

### 4.2 Python — 按状态分类统计

```python
def analyze_usage(redemptions):
    now = int(time.time())
    used, unused, disabled, expired = [], [], [], []

    for r in redemptions:
        if r["status"] == 3:
            used.append(r)
        elif r["status"] == 2:
            disabled.append(r)
        elif r["expired_time"] != 0 and r["expired_time"] < now:
            expired.append(r)
        else:
            unused.append(r)

    total_redeemed_quota = sum(r["quota"] for r in used)

    print(f"总数: {len(redemptions)}")
    print(f"已使用: {len(used)}, 总兑换额度: {total_redeemed_quota}")
    print(f"未使用: {len(unused)}")
    print(f"已禁用: {len(disabled)}")
    print(f"已过期: {len(expired)}")

    return {"used": used, "unused": unused, "disabled": disabled, "expired": expired}
```

### 4.3 Python — 查看兑换码使用详情

```python
def get_usage_detail(redemption):
    if redemption["status"] != 3:
        return None
    return {
        "code_name": redemption["name"],
        "quota": redemption["quota"],
        "used_by_user_id": redemption["used_user_id"],
        "used_at": redemption["redeemed_time"],
    }
```

### 4.4 cURL 示例

```bash
# 获取第1页，每页20条
curl -H "Authorization: YOUR_TOKEN" \
  "https://your-domain.com/api/redemption/?p=1&page_size=20"

# 按名称搜索
curl -H "Authorization: YOUR_TOKEN" \
  "https://your-domain.com/api/redemption/search?keyword=活动&p=1&page_size=20"

# 查询单个兑换码
curl -H "Authorization: YOUR_TOKEN" \
  "https://your-domain.com/api/redemption/42"
```

---

## 5. 当前 API 局限性

| 能力 | 是否支持 | 备注 |
|------|----------|------|
| 查看兑换人 ID | ✅ | `used_user_id` 字段 |
| 查看兑换时间 | ✅ | `redeemed_time` 字段 |
| 按状态筛选 | ❌ | 需客户端过滤 |
| 按兑换人筛选 | ❌ | 需客户端过滤 |
| 按时间范围筛选 | ❌ | 需客户端过滤 |
| 聚合统计 | ❌ | 需客户端计算 |
| 兑换人用户名/邮箱 | ❌ | 仅返回 user_id，需二次调用 `GET /api/user/{id}` |
