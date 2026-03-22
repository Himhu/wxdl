# 兑换码管理接口文档

> 版本: v1.0
> 更新日期: 2026-03-14

---

## 目录

1. [概述](#1-概述)
2. [认证方式](#2-认证方式)
3. [通用响应格式](#3-通用响应格式)
4. [分页参数](#4-分页参数)
5. [创建兑换码（批量）](#5-创建兑换码批量)
6. [获取兑换码列表](#6-获取兑换码列表)
7. [获取单个兑换码](#7-获取单个兑换码)
8. [搜索兑换码](#8-搜索兑换码)
9. [更新兑换码](#9-更新兑换码)
10. [删除单个兑换码](#10-删除单个兑换码)
11. [清理无效兑换码](#11-清理无效兑换码)
12. [用户兑换接口](#12-用户兑换接口)
13. [兑换码对象字段说明](#13-兑换码对象字段说明)
14. [错误码参考](#14-错误码参考)
15. [完整对接流程示例](#15-完整对接流程示例)

---

## 1. 概述

兑换码（Redemption Code）用于向用户发放额度。管理员可批量创建兑换码，用户输入兑换码即可充值对应额度到账户。

接口分为两类：

| 类别 | 认证要求 | 路由前缀 | 说明 |
|---|---|---|---|
| **管理端** | AdminAuth（管理员权限） | `/api/redemption/` | 创建、查看、更新、删除兑换码 |
| **用户端** | UserAuth（普通用户权限） | `/api/user/self/topup` | 使用兑换码充值额度 |

---

## 2. 认证方式

### 管理端接口（AdminAuth）

所有 `/api/redemption/*` 接口均需要**管理员权限**。支持两种认证方式，**二选一**，但都必须携带 `New-Api-User` 请求头。

#### 方式 A：Session Cookie 认证

适用于浏览器端，管理员登录后自动携带。

```
Cookie: session=<session_id>
New-Api-User: <管理员用户ID>
```

#### 方式 B：Access Token 认证

适用于后端服务或脚本对接。

```
Authorization: Bearer <access_token>
New-Api-User: <管理员用户ID>
```

> **重要**：
> - `New-Api-User` 请求头**必须**提供，且值必须与当前认证用户的 ID 一致
> - 认证用户的角色必须 ≥ 管理员（role ≥ 10），否则返回 403

### 用户端接口（UserAuth）

`POST /api/user/self/topup` 只需要普通用户权限，认证方式与管理端相同（Session 或 Access Token + `New-Api-User` 头），但不要求管理员角色。

### 认证失败响应

| 场景 | HTTP 状态码 | message |
|---|---|---|
| 未登录且未提供 access token | 401 | 无权进行此操作，未登录且未提供 access token |
| access token 无效 | 200 | 无权进行此操作，access token 无效 |
| 缺少 New-Api-User 头 | 401 | 无权进行此操作，未提供 New-Api-User |
| New-Api-User 格式错误 | 401 | 无权进行此操作，New-Api-User 格式错误 |
| New-Api-User 与登录用户不匹配 | 401 | 无权进行此操作，New-Api-User 与登录用户不匹配 |
| 用户角色不足（非管理员访问管理端接口） | 403 | 无权进行此操作，权限不足 |

---

## 3. 通用响应格式

所有接口均返回 JSON，外层结构统一为：

**成功：**
```json
{
  "success": true,
  "message": "",
  "data": ...
}
```

**失败：**
```json
{
  "success": false,
  "message": "错误描述"
}
```

> 注意：大部分业务错误仍返回 **HTTP 200**，需通过 `success` 字段判断是否成功。仅认证/权限失败返回 401/403。

---

## 4. 分页参数

以下接口支持分页：`GET /api/redemption/`、`GET /api/redemption/search`

### 请求参数（Query String）

| 参数 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `p` | int | 1 | 页码（从 1 开始） |
| `page_size` | int | 系统默认值 | 每页数量，最大 100。别名：`ps`、`size` |

> **注意**：页码参数是 **`p`**，不是 `page`。

### 分页响应结构

```json
{
  "success": true,
  "message": "",
  "data": {
    "page": 1,
    "page_size": 10,
    "total": 50,
    "items": [ ... ]
  }
}
```

---

## 5. 创建兑换码（批量）

此接口支持**一次性批量创建**多个兑换码，每个兑换码共享相同的名称、额度和过期时间，但拥有独立的 key。

### 请求

```
POST /api/redemption/
Content-Type: application/json
```

### 请求体

```json
{
  "name": "新用户福利",
  "count": 10,
  "quota": 500000,
  "expired_time": 1767225600
}
```

### 请求字段说明

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|---|---|---|---|---|
| `name` | string | 是 | 1 ~ 20 个字符（UTF-8） | 兑换码名称，用于管理标识 |
| `count` | int | 是 | 1 ~ 100 | 批量创建的数量 |
| `quota` | int | 是 | 正整数 | 每个兑换码的充值额度（内部额度单位） |
| `expired_time` | int64 | 否 | `0` 或 大于当前时间的时间戳 | 过期时间（Unix 秒级时间戳）。`0` 表示永不过期 |

### 成功响应

```json
{
  "success": true,
  "message": "",
  "data": [
    "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    "c3d4e5f6-a7b8-9012-cdef-123456789012"
  ]
}
```

> **与 Token 接口的关键区别**：创建兑换码接口**直接返回所有生成的 key**（UUID 格式的字符串数组），无需额外调用列表接口获取。请在创建后立即保存这些 key，后续只能通过列表/详情接口查看。

### 部分成功场景

如果批量创建过程中某个兑换码插入失败，接口会**立即停止创建**并返回已成功创建的 key 列表：

```json
{
  "success": false,
  "message": "兑换码创建失败",
  "data": [
    "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  ]
}
```

> `data` 中包含已成功创建的 key。如果全部失败，`data` 为 `null`。

### 失败响应示例

```json
{"success": false, "message": "兑换码名称长度必须为1到20之间"}
```
```json
{"success": false, "message": "兑换码创建数量必须为正整数"}
```
```json
{"success": false, "message": "兑换码创建数量不能超过100"}
```
```json
{"success": false, "message": "兑换码过期时间不能早于当前时间"}
```

### curl 示例

```bash
# 批量创建 5 个兑换码，每个 100万额度，永不过期
curl -X POST 'https://your-domain.com/api/redemption/' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>' \
  -d '{
    "name": "新用户福利",
    "count": 5,
    "quota": 1000000,
    "expired_time": 0
  }'
```

---

## 6. 获取兑换码列表

### 请求

```
GET /api/redemption/?p=1&page_size=10
```

### 成功响应

```json
{
  "success": true,
  "message": "",
  "data": {
    "page": 1,
    "page_size": 10,
    "total": 50,
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "key": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "status": 1,
        "name": "新用户福利",
        "quota": 500000,
        "created_time": 1773000000,
        "redeemed_time": 0,
        "used_user_id": 0,
        "expired_time": 0
      },
      {
        "id": 2,
        "user_id": 1,
        "key": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
        "status": 3,
        "name": "新用户福利",
        "quota": 500000,
        "created_time": 1773000000,
        "redeemed_time": 1773100000,
        "used_user_id": 5,
        "expired_time": 0
      }
    ]
  }
}
```

> 列表按 ID **倒序**排列，最新创建的在最前面。

### curl 示例

```bash
curl -X GET 'https://your-domain.com/api/redemption/?p=1&page_size=20' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>'
```

---

## 7. 获取单个兑换码

### 请求

```
GET /api/redemption/:id
```

### 路径参数

| 参数 | 类型 | 说明 |
|---|---|---|
| `id` | int | 兑换码 ID |

### 成功响应

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 1,
    "user_id": 1,
    "key": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "status": 1,
    "name": "新用户福利",
    "quota": 500000,
    "created_time": 1773000000,
    "redeemed_time": 0,
    "used_user_id": 0,
    "expired_time": 0
  }
}
```

### curl 示例

```bash
curl -X GET 'https://your-domain.com/api/redemption/1' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>'
```

---

## 8. 搜索兑换码

### 请求

```
GET /api/redemption/search?keyword=福利&p=1&page_size=10
```

### 查询参数

| 参数 | 类型 | 说明 |
|---|---|---|
| `keyword` | string | 搜索关键词。支持两种匹配方式（自动识别）：|
| | | - 纯数字：按 **ID 精确匹配** 或 **名称前缀匹配** |
| | | - 非数字：仅按 **名称前缀匹配** |
| `p` | int | 页码 |
| `page_size` | int | 每页条数 |

> 搜索按名称**前缀匹配**（`name LIKE 'keyword%'`），不支持中间或尾部模糊匹配。

### 成功响应

与 [获取兑换码列表](#6-获取兑换码列表) 格式一致。

### curl 示例

```bash
# 按名称搜索
curl -X GET 'https://your-domain.com/api/redemption/search?keyword=新用户' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>'

# 按 ID 搜索
curl -X GET 'https://your-domain.com/api/redemption/search?keyword=42' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>'
```

---

## 9. 更新兑换码

### 请求

```
PUT /api/redemption/
Content-Type: application/json
```

> 通过**请求体**中的 `id` 字段定位要更新的兑换码（不是路径参数）。

### 模式一：全量更新（默认）

更新名称、额度和过期时间。

```json
{
  "id": 1,
  "name": "春节特惠",
  "quota": 1000000,
  "expired_time": 1767225600
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int | **必填**，要更新的兑换码 ID |
| `name` | string | 新的兑换码名称 |
| `quota` | int | 新的额度值 |
| `expired_time` | int64 | 新的过期时间。`0` 表示永不过期，非 `0` 必须晚于当前时间 |

> **注意**：`key`、`user_id`、`created_time` 等字段不可通过此接口修改。

### 模式二：仅更新状态

```
PUT /api/redemption/?status_only=1
```

```json
{
  "id": 1,
  "status": 2
}
```

当 URL 携带 `status_only` 参数（任意非空值）时，只更新 `status` 字段。

### 成功响应

返回更新后的完整兑换码对象：

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 1,
    "user_id": 1,
    "key": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "status": 1,
    "name": "春节特惠",
    "quota": 1000000,
    "created_time": 1773000000,
    "redeemed_time": 0,
    "used_user_id": 0,
    "expired_time": 1767225600
  }
}
```

### curl 示例

```bash
# 全量更新
curl -X PUT 'https://your-domain.com/api/redemption/' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>' \
  -d '{
    "id": 1,
    "name": "春节特惠",
    "quota": 1000000,
    "expired_time": 0
  }'

# 仅禁用兑换码
curl -X PUT 'https://your-domain.com/api/redemption/?status_only=1' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>' \
  -d '{"id": 1, "status": 2}'
```

---

## 10. 删除单个兑换码

### 请求

```
DELETE /api/redemption/:id
```

### 路径参数

| 参数 | 类型 | 说明 |
|---|---|---|
| `id` | int | 兑换码 ID |

### 成功响应

```json
{
  "success": true,
  "message": ""
}
```

> 删除为**软删除**（设置 `deleted_at`），不会从数据库中物理删除。

### curl 示例

```bash
curl -X DELETE 'https://your-domain.com/api/redemption/1' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>'
```

---

## 11. 清理无效兑换码

一次性删除所有"无效"兑换码。

### 请求

```
DELETE /api/redemption/invalid
```

> **无需传递任何参数**，此接口会自动识别并删除所有符合以下条件的兑换码。

### 清理范围

以下兑换码会被删除：

| 条件 | 说明 |
|---|---|
| `status = 3`（已使用） | 已被用户兑换的兑换码 |
| `status = 2`（已禁用） | 被管理员手动禁用的兑换码 |
| `status = 1` 且 `expired_time ≠ 0` 且 `expired_time < 当前时间` | 已过期但状态仍为启用的兑换码 |

### 成功响应

```json
{
  "success": true,
  "message": "",
  "data": 15
}
```

`data` 为实际删除的兑换码数量。

> **注意**：此操作为**软删除**，但影响范围较大。建议在操作前先通过列表/搜索接口确认将被清理的兑换码数量。

### curl 示例

```bash
curl -X DELETE 'https://your-domain.com/api/redemption/invalid' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>'
```

---

## 12. 用户兑换接口

此接口供**普通用户**使用兑换码充值额度，不需要管理员权限。

### 请求

```
POST /api/user/self/topup
Content-Type: application/json
```

> 此接口有**频率限制**（CriticalRateLimit 中间件）和**并发锁**（同一用户同时只能发起一次兑换请求）。

### 请求体

```json
{
  "key": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `key` | string | 是 | 兑换码（UUID 格式，由管理员创建时生成） |

### 成功响应

```json
{
  "success": true,
  "message": "",
  "data": 500000
}
```

`data` 为本次充值的额度数值。

### 失败响应

| 场景 | message |
|---|---|
| 兑换码不存在 | 兑换失败 |
| 兑换码已被使用 | 兑换失败 |
| 兑换码已过期 | 兑换失败 |
| 用户正在兑换中（并发请求） | 正在处理中，请勿重复提交 |

> **设计说明**：出于安全考虑，兑换失败时不区分具体原因（不存在 / 已使用 / 已过期），统一返回 "兑换失败"，防止通过错误信息枚举有效兑换码。

### 兑换流程（内部逻辑）

```
用户提交兑换码
  → 获取用户级别并发锁（同一用户只能同时兑换一次）
  → 数据库行锁（SELECT FOR UPDATE）查找兑换码
  → 校验状态是否为启用（status = 1）
  → 校验是否过期（expired_time = 0 或 > 当前时间）
  → 增加用户额度（UPDATE users SET quota = quota + ?)
  → 标记兑换码为已使用（status = 3），记录兑换时间和兑换用户
  → 提交事务
  → 写入充值日志
```

### curl 示例

```bash
curl -X POST 'https://your-domain.com/api/user/self/topup' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <user_access_token>' \
  -H 'New-Api-User: <user_id>' \
  -d '{
    "key": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }'
```

---

## 13. 兑换码对象字段说明

以下是兑换码对象的完整字段列表，适用于列表、详情、更新响应中返回的对象。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int | 兑换码唯一标识（数据库自增） |
| `user_id` | int | 创建该兑换码的管理员用户 ID |
| `key` | string | 兑换码密钥（32 位 UUID 格式，如 `a1b2c3d4-e5f6-7890-abcd-ef1234567890`） |
| `status` | int | 状态码（见下表） |
| `name` | string | 兑换码名称（管理标识用途） |
| `quota` | int | 兑换码面额（内部额度单位） |
| `created_time` | int64 | 创建时间（Unix 秒级时间戳） |
| `redeemed_time` | int64 | 兑换时间（Unix 秒级时间戳）。未兑换时为 `0` |
| `used_user_id` | int | 兑换该码的用户 ID。未兑换时为 `0` |
| `expired_time` | int64 | 过期时间（Unix 秒级时间戳）。`0` 表示永不过期 |

### 状态码说明

| 值 | 含义 | 说明 |
|---|---|---|
| `1` | 启用（enabled） | 可被用户兑换 |
| `2` | 禁用（disabled） | 被管理员手动禁用，不可兑换 |
| `3` | 已使用（used） | 已被用户兑换，不可再次使用 |

> **注意**：兑换码是**一次性**的。每个兑换码只能被一个用户兑换一次，兑换后自动变为"已使用"状态。

---

## 14. 错误码参考

### 管理端错误

| 场景 | message |
|---|---|
| 兑换码名称为空或超过 20 字符 | 兑换码名称长度必须为1到20之间 |
| 创建数量 ≤ 0 | 兑换码创建数量必须为正整数 |
| 创建数量 > 100 | 兑换码创建数量不能超过100 |
| 过期时间早于当前时间 | 兑换码过期时间不能早于当前时间 |
| 数据库插入失败 | 兑换码创建失败 |
| 兑换码 ID 不存在 | record not found |
| ID 参数格式错误 | 参数错误 |

### 用户兑换错误

| 场景 | message |
|---|---|
| 兑换码不存在 / 已使用 / 已过期 | 兑换失败 |
| 并发重复提交 | 正在处理中，请勿重复提交 |
| 请求体格式错误 | 参数错误 |

---

## 15. 完整对接流程示例

### 场景 A：管理端 — 批量创建并分发兑换码

```bash
# 步骤 1：创建 10 个兑换码，每个 50万额度，2026年底过期
curl -X POST 'https://your-domain.com/api/redemption/' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>' \
  -d '{
    "name": "2026春季促销",
    "count": 10,
    "quota": 500000,
    "expired_time": 1798761600
  }'

# 响应 — 立即获得所有兑换码 key
# {
#   "success": true,
#   "data": [
#     "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
#     "b2c3d4e5-f6a7-8901-bcde-f12345678901",
#     ...共10个
#   ]
# }

# 步骤 2：将兑换码分发给用户（通过邮件、短信、页面展示等方式）

# 步骤 3：查看兑换码使用情况
curl -X GET 'https://your-domain.com/api/redemption/search?keyword=2026春季促销' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>'

# 步骤 4（可选）：清理已使用/已过期的兑换码
curl -X DELETE 'https://your-domain.com/api/redemption/invalid' \
  -H 'Authorization: Bearer <admin_access_token>' \
  -H 'New-Api-User: <admin_user_id>'
```

### 场景 B：用户端 — 兑换码充值

```bash
# 用户输入收到的兑换码
curl -X POST 'https://your-domain.com/api/user/self/topup' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <user_access_token>' \
  -H 'New-Api-User: <user_id>' \
  -d '{
    "key": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }'

# 成功响应
# {
#   "success": true,
#   "data": 500000
# }
# 用户账户额度增加 500000
```

---

## 附录：接口速查表

| 方法 | 路径 | 认证 | 说明 |
|---|---|---|---|
| `POST` | `/api/redemption/` | AdminAuth | 批量创建兑换码 |
| `GET` | `/api/redemption/?p=1&page_size=10` | AdminAuth | 获取兑换码列表（分页） |
| `GET` | `/api/redemption/:id` | AdminAuth | 获取单个兑换码详情 |
| `GET` | `/api/redemption/search?keyword=xxx` | AdminAuth | 按名称/ID 搜索兑换码 |
| `PUT` | `/api/redemption/` | AdminAuth | 更新兑换码（全量） |
| `PUT` | `/api/redemption/?status_only=1` | AdminAuth | 仅更新兑换码状态 |
| `DELETE` | `/api/redemption/:id` | AdminAuth | 删除单个兑换码 |
| `DELETE` | `/api/redemption/invalid` | AdminAuth | 清理所有无效兑换码 |
| `POST` | `/api/user/self/topup` | UserAuth | 用户兑换码充值 |
