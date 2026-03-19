# 代理商管理系统 - 完整部署指南

## 系统概述

这是一个完整的代理商管理系统，支持：
- ✅ 微信授权登录 + 账号密码登录
- ✅ 多级代理商管理（最多3级）
- ✅ 上级给下级充值积分（审批制）
- ✅ 卡密生成和管理
- ✅ 完整的日志系统

## 核心业务逻辑

### 1. 积分充值流程
```
下级代理 → 申请充值 → 上级代理审批 → 通过后扣除上级积分 → 下级获得积分
```

### 2. 卡密生成流程
```
代理商 → 消耗积分 → 生成卡密 → 卡密可用于激活/授权
```

### 3. 代理商层级
```
一级代理（顶级）
  └── 二级代理
        └── 三级代理
```

## 快速部署

### 第一步：部署后端

```bash
cd backend

# 1. 安装 Go 依赖
go mod download

# 2. 配置环境变量
cp .env.example .env
nano .env  # 编辑配置

# 必须配置的项目：
# - DB_PASSWORD: 你的 MySQL 密码
# - JWT_SECRET: 随机生成的密钥
# - WECHAT_APPID: 微信小程序 AppID
# - WECHAT_SECRET: 微信小程序 Secret

# 3. 创建数据库
mysql -u root -p -e "CREATE DATABASE agent_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 4. 导入数据库表结构
mysql -u root -p agent_system < scripts/init.sql

# 5. 启动后端服务
go run cmd/server/main.go
# 或使用启动脚本
./start.sh
```

后端将在 `http://localhost:8080` 启动。

### 第二步：配置前端

```bash
cd miniprogram

# 1. 修改 API 地址
# 编辑 api/request.js
# 将 baseURL 改为你的后端地址
```

在微信开发者工具中：
1. 打开 `miniprogram` 目录
2. 配置 AppID
3. 编译运行

### 第三步：初始化数据

```sql
-- 创建测试站点
INSERT INTO stations (name, code, status) VALUES ('测试站点', 'TEST001', 1);

-- 创建一级代理（初始积分 10000）
INSERT INTO agents (username, password, real_name, level, balance, status, station_id)
VALUES ('admin', '$2a$10$...', '系统管理员', 1, 10000, 1, 1);
```

## API 接口文档

### 认证接口

#### 1. 账号密码登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "123456",
  "stationId": 1
}

Response:
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGc...",
    "userInfo": {
      "id": 1,
      "username": "admin",
      "level": 1,
      "balance": 10000
    }
  }
}
```

#### 2. 微信授权登录
```http
POST /api/auth/wechat/login
Content-Type: application/json

{
  "code": "wx_code_from_wx_login",
  "stationId": 1
}
```

### 积分管理接口

#### 1. 申请充值（下级向上级申请）
```http
POST /api/points/recharge/apply
Authorization: Bearer {token}
Content-Type: application/json

{
  "amount": 1000,
  "remark": "申请充值1000积分"
}
```

#### 2. 查看待审批列表（上级查看）
```http
GET /api/points/recharge/pending?page=1&pageSize=20
Authorization: Bearer {token}
```

#### 3. 审批通过
```http
POST /api/points/recharge/approve/1
Authorization: Bearer {token}
```

#### 4. 审批拒绝
```http
POST /api/points/recharge/reject/1
Authorization: Bearer {token}
Content-Type: application/json

{
  "rejectReason": "积分不足"
}
```

### 卡密管理接口

#### 1. 生成卡密
```http
POST /api/cards
Authorization: Bearer {token}
Content-Type: application/json

{
  "count": 10
}
```

#### 2. 查询卡密列表
```http
GET /api/cards?page=1&pageSize=20&status=1
Authorization: Bearer {token}
```

### 代理商管理接口

#### 1. 创建下级代理
```http
POST /api/agents
Authorization: Bearer {token}
Content-Type: application/json

{
  "username": "agent002",
  "password": "123456",
  "realName": "张三",
  "phone": "13800138000"
}
```

#### 2. 查询下级代理列表
```http
GET /api/agents?page=1&pageSize=20
Authorization: Bearer {token}
```

## 业务流程示例

### 场景1：一级代理给二级代理充值

1. **二级代理申请充值**
```bash
curl -X POST http://localhost:8080/api/points/recharge/apply \
  -H "Authorization: Bearer {二级代理token}" \
  -H "Content-Type: application/json" \
  -d '{"amount": 1000, "remark": "申请充值"}'
```

2. **一级代理查看待审批**
```bash
curl -X GET http://localhost:8080/api/points/recharge/pending \
  -H "Authorization: Bearer {一级代理token}"
```

3. **一级代理审批通过**
```bash
curl -X POST http://localhost:8080/api/points/recharge/approve/1 \
  -H "Authorization: Bearer {一级代理token}"
```

4. **系统自动执行**
   - 扣除一级代理 1000 积分
   - 增加二级代理 1000 积分
   - 记录积分流水

### 场景2：代理商生成卡密

```bash
curl -X POST http://localhost:8080/api/cards \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"count": 10}'
```

系统会：
- 生成 10 个卡密
- 扣除对应积分
- 记录操作日志

## 数据库表说明

### 核心表
- `agents` - 代理商表
- `cards` - 卡密表
- `points_records` - 积分流水表
- `recharge_requests` - 充值申请表
- `operation_logs` - 操作日志表
- `login_logs` - 登录日志表
- `stations` - 站点表

## 安全建议

1. **生产环境必须修改**：
   - JWT_SECRET（使用强随机密钥）
   - 数据库密码
   - 管理员初始密码

2. **启用 HTTPS**：
   - 使用 Nginx 反向代理
   - 配置 SSL 证书

3. **数据库安全**：
   - 定期备份
   - 限制远程访问
   - 使用强密码

4. **日志管理**：
   - 定期清理旧日志
   - 监控异常操作

## 常见问题

### Q1: 如何重置管理员密码？
```sql
UPDATE agents SET password = '$2a$10$...' WHERE username = 'admin';
```

### Q2: 如何查看某个代理的所有下级？
```sql
SELECT * FROM agents WHERE parent_id = {代理ID};
```

### Q3: 如何查看积分流水？
```sql
SELECT * FROM points_records WHERE agent_id = {代理ID} ORDER BY created_at DESC;
```

## 技术支持

- 后端文档：[backend/README.md](backend/README.md)
- 项目总览：[PROJECT_OVERVIEW.md](PROJECT_OVERVIEW.md)

## License

MIT
