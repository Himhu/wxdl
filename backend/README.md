# 代理商管理系统后端

基于 Go + Gin + MySQL 开发的代理商管理系统后端服务。

## 技术栈

- **Go 1.21+**
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **MySQL 8.0+** - 数据库
- **JWT** - 身份认证
- **Bcrypt** - 密码加密

## 功能模块

### 1. 用户认证系统
- ✅ 账号密码登录
- ✅ 微信授权登录
- ✅ 微信账号绑定
- ✅ JWT Token 管理
- ✅ 登录日志记录

### 2. 代理商管理
- ✅ 创建下级代理（最多3级）
- ✅ 查询代理列表
- ✅ 更新代理信息
- ✅ 禁用/启用代理

### 3. 卡密管理
- ✅ 批量生成卡密
- ✅ 查询卡密列表
- ✅ 销毁卡密
- ✅ 卡密统计

### 4. 积分系统
- ✅ 充值申请与上级审批
- ✅ 积分消费
- ✅ 积分流水查询
- ✅ 积分统计

### 5. 日志系统
- ✅ 操作日志
- ✅ 登录日志

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 8.0+
- Git

### 2. 安装依赖

```bash
cd backend
go mod download
```

### 3. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库和微信小程序信息
```

### 4. 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE agent_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入表结构
mysql -u root -p agent_system < scripts/init.sql
```

### 5. 运行服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## API 文档

### 认证相关

#### 1. 账号密码登录
```
POST /api/auth/login
Content-Type: application/json

{
  "username": "agent001",
  "password": "123456",
  "stationId": 1
}
```

#### 2. 微信授权登录
```
POST /api/auth/wechat/login
Content-Type: application/json

{
  "code": "wx_login_code",
  "stationId": 1
}
```

#### 3. 微信账号绑定
```
POST /api/auth/wechat/bind
Content-Type: application/json

{
  "username": "agent001",
  "password": "123456",
  "openId": "wx_open_id",
  "stationId": 1
}
```

#### 4. 获取当前用户信息
```
GET /api/auth/me
Authorization: Bearer {token}
```

### 卡密管理

#### 1. 创建卡密
```
POST /api/cards
Authorization: Bearer {token}
Content-Type: application/json

{
  "count": 10
}
```

#### 2. 查询卡密列表
```
GET /api/cards?page=1&pageSize=20&status=1
Authorization: Bearer {token}
```

#### 3. 销毁卡密
```
DELETE /api/cards/:id
Authorization: Bearer {token}
```

### 代理商管理

#### 1. 创建下级代理
```
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
```
GET /api/agents?page=1&pageSize=20
Authorization: Bearer {token}
```

### 积分系统

#### 1. 查询积分余额
```
GET /api/points/balance
Authorization: Bearer {token}
```

#### 2. 提交充值申请
```
POST /api/points/recharge/apply
Authorization: Bearer {token}
Content-Type: application/json

{
  "amount": 1000,
  "paymentMethod": "bank",
  "remark": "申请充值1000积分"
}
```

#### 3. 查看待审批列表
```
GET /api/points/recharge/pending?page=1&pageSize=20
Authorization: Bearer {token}
```

#### 4. 审批通过
```
POST /api/points/recharge/approve/:id
Authorization: Bearer {token}
```

#### 5. 审批拒绝
```
POST /api/points/recharge/reject/:id
Authorization: Bearer {token}
```

#### 6. 充值历史
```
GET /api/points/recharge/history?page=1&pageSize=20&status=1
Authorization: Bearer {token}
```

#### 7. 积分流水查询
```
GET /api/points/records?page=1&pageSize=20&type=1
Authorization: Bearer {token}
```

## 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── config/                  # 配置管理
│   ├── handler/                 # 请求处理器
│   │   ├── auth_handler.go
│   │   ├── card_handler.go
│   │   ├── agent_handler.go
│   │   └── points_handler.go
│   ├── service/                 # 业务逻辑层
│   │   ├── auth_service.go
│   │   ├── card_service.go
│   │   ├── agent_service.go
│   │   └── points_service.go
│   ├── repository/              # 数据访问层
│   │   ├── agent_repository.go
│   │   ├── card_repository.go
│   │   └── points_repository.go
│   ├── model/                   # 数据模型
│   │   └── models.go
│   ├── middleware/              # 中间件
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logger.go
│   └── utils/                   # 工具函数
│       ├── jwt.go
│       ├── wechat.go
│       └── response.go
├── scripts/
│   └── init.sql                 # 数据库初始化脚本
├── .env.example                 # 环境变量示例
├── go.mod
├── go.sum
└── README.md
```

## 数据库设计

### 核心表

- **agents** - 代理商表
- **cards** - 卡密表
- **balance_logs** - 积分记录表
- **recharge_requests** - 充值申请表
- **operation_logs** - 操作日志表
- **login_logs** - 登录日志表
- **stations** - 站点表

详细表结构请查看 `scripts/init.sql`。

## 开发说明

### 添加新的 API 接口

1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/repository/` 中实现数据访问
3. 在 `internal/service/` 中实现业务逻辑
4. 在 `internal/handler/` 中实现请求处理
5. 在 `cmd/server/main.go` 中注册路由

### 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 官方编码规范
- 添加必要的注释
- 错误处理要完整

## 部署

### 使用 Docker 部署

```bash
# 构建镜像
docker build -t agent-system-backend .

# 运行容器
docker run -d \
  --name agent-backend \
  -p 8080:8080 \
  --env-file .env \
  agent-system-backend
```

### 使用 systemd 部署

```bash
# 编译
go build -o agent-system cmd/server/main.go

# 创建 systemd 服务
sudo cp agent-system.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable agent-system
sudo systemctl start agent-system
```

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service/...

# 查看测试覆盖率
go test -cover ./...
```

## 常见问题

### 1. 数据库连接失败
检查 `.env` 文件中的数据库配置是否正确。

### 2. JWT Token 验证失败
确保 `JWT_SECRET` 配置正确，且 Token 未过期。

### 3. 微信授权失败
检查微信小程序的 AppID 和 Secret 是否配置正确。

## License

MIT
