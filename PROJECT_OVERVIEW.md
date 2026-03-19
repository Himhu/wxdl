# 代理商管理系统 - 项目总览

## 项目简介

这是一个完整的代理商管理系统，包含微信小程序前端和 Go 后端。

## 技术栈

### 前端（微信小程序）
- 微信小程序原生开发
- 微信授权登录
- 现代化 UI 设计

### 后端（Go）
- Go 1.21+
- Gin Web 框架
- GORM ORM
- MySQL 8.0+
- JWT 认证

## 项目结构

```
微信小程序/
├── miniprogram/              # 小程序前端
│   ├── pages/               # 页面
│   │   ├── common/         # 公共页面（登录、绑定）
│   │   └── tabbar/         # 主页面（工作台、卡密、个人中心）
│   ├── components/         # 组件
│   ├── api/                # API 封装
│   ├── utils/              # 工具函数
│   └── styles/             # 全局样式
│
└── backend/                 # Go 后端
    ├── cmd/                # 程序入口
    ├── internal/           # 内部代码
    │   ├── handler/       # 请求处理
    │   ├── service/       # 业务逻辑
    │   ├── repository/    # 数据访问
    │   ├── model/         # 数据模型
    │   ├── middleware/    # 中间件
    │   └── utils/         # 工具函数
    └── scripts/           # 脚本文件
```

## 核心功能

### ✅ 用户认证
- 账号密码登录
- 微信授权登录
- 微信账号绑定
- JWT Token 管理

### ✅ 代理商管理
- 多级代理体系（最多3级）
- 创建下级代理
- 查询代理列表
- 更新代理信息

### ✅ 卡密管理
- 批量生成卡密
- 查询卡密列表
- 销毁卡密
- 卡密统计

### ✅ 积分系统
- 积分充值
- 积分消费
- 积分流水查询
- 积分统计

### ✅ 日志系统
- 操作日志
- 登录日志

## 快速开始

### 1. 后端启动

```bash
cd backend

# 安装依赖
go mod download

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件

# 初始化数据库
mysql -u root -p -e "CREATE DATABASE agent_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p agent_system < scripts/init.sql

# 启动服务
./start.sh
# 或
go run cmd/server/main.go
```

后端将在 `http://localhost:8080` 启动。

### 2. 前端配置

```bash
cd miniprogram

# 修改 API 地址
# 编辑 api/request.js，将 baseURL 改为后端地址
```

在微信开发者工具中打开 `miniprogram` 目录。

### 3. 测试账号

```
账号: agent001
密码: 123456
站点ID: 1
```

## API 接口

### 认证相关
- `POST /api/auth/login` - 账号密码登录
- `POST /api/auth/wechat/login` - 微信授权登录
- `POST /api/auth/wechat/bind` - 微信账号绑定
- `GET /api/auth/me` - 获取当前用户信息

### 卡密管理
- `POST /api/cards` - 创建卡密
- `GET /api/cards` - 查询卡密列表
- `DELETE /api/cards/:id` - 销毁卡密
- `GET /api/cards/stats` - 卡密统计

### 代理商管理
- `POST /api/agents` - 创建下级代理
- `GET /api/agents` - 查询代理列表
- `GET /api/agents/:id` - 查询代理详情
- `PUT /api/agents/:id` - 更新代理信息

### 积分系统
- `GET /api/points/balance` - 查询积分余额
- `POST /api/points/recharge` - 积分充值
- `GET /api/points/records` - 积分流水查询
- `GET /api/points/stats` - 积分统计

## 数据库设计

### 核心表
- `agents` - 代理商表
- `cards` - 卡密表
- `points_records` - 积分记录表
- `operation_logs` - 操作日志表
- `login_logs` - 登录日志表
- `stations` - 站点表

详细设计请查看 `backend/scripts/init.sql`。

## UI 设计

### 配色方案
- 主色：#2563EB（科技蓝）
- 成功：#16A34A（绿色）
- 警告：#D97706（橙色）
- 危险：#DC2626（红色）
- 微信：#07C160（微信绿）

### 设计风格
- 简约现代
- 卡片式布局
- 渐变背景
- 圆角设计

## 开发团队

本项目由 Claude Code 协同 Codex 和 Gemini 共同开发完成。

## 后续优化建议

1. **前端优化**
   - 添加下拉刷新
   - 添加骨架屏
   - 优化加载动画

2. **后端优化**
   - 添加 Redis 缓存
   - 添加接口限流
   - 完善错误处理

3. **功能扩展**
   - 添加数据统计图表
   - 添加消息推送
   - 添加导出功能

## License

MIT
