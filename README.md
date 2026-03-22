# 微信代理小程序

微信小程序 + 管理后台 + Go 后端的代理商管理系统。

## 项目结构

```
├── backend/            # Go 后端（统一 API，服务小程序和管理端）
├── admin-frontend/     # 管理后台（React + Ant Design）
├── miniprogram/        # 微信小程序
└── docs/               # 项目文档
```

## 快速启动

### 1. 后端

```bash
cd backend
cp configs/config.yaml.example configs/config.yaml  # 首次需配置数据库和微信参数
go run cmd/server/main.go
```

默认监听 `0.0.0.0:8080`

### 2. 管理端

```bash
cd admin-frontend
npm install
npm run dev
```

默认访问 `http://localhost:5174`，管理员账号 `admin / admin123`

### 3. 小程序

用微信开发者工具打开 `miniprogram/` 目录，配置 `config/index.js` 中的 `API_BASE_URL` 指向后端地址。

## 核心功能

| 模块 | 说明 |
|------|------|
| 微信登录 | 一键登录自动注册，代理申请由管理员审批 |
| 代理邀请 | 生成邀请码/小程序码海报，扫码自动绑定邀请关系 |
| 卡密管理 | 卡密生成、分发、核销、作废 |
| 积分体系 | 充值申请、审批、余额管理 |
| 数据转移 | 旧站余额查询、确认转移、旧站账号禁用 |
| 对象存储 | 支持雨云 ROS / 阿里云 OSS / MinIO（S3 兼容） |
| 管理后台 | 数据概览、用户管理、卡密管理、系统设置 |

## 技术栈

- **后端**：Go + Gin + GORM + MySQL + JWT
- **管理端**：React + TypeScript + Ant Design + Vite
- **小程序**：原生微信小程序 + MobX

## 文档

详细文档见 [docs/](docs/) 目录。
