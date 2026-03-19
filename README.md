# 代理站管理微信小程序

## 项目简介
这是一个用于代理站管理的微信小程序，支持多站点切换、卡密管理、积分充值等功能。

## 功能特性
- ✅ 双站点独立管理（站点A、站点B）
- ✅ 卡密管理（创建、查询、销毁）
- ✅ 批量创建卡密
- ✅ 多级代理系统
- ✅ 线下充值审核
- ✅ 完整操作日志
- ✅ 站点快速切换

## 技术栈
- 微信小程序原生框架
- MobX 状态管理
- 模块化 API 封装

## 目录结构
```
miniprogram/
├── api/              # API 接口层
│   ├── request.js    # 请求封装
│   ├── card.js       # 卡密接口
│   ├── finance.js    # 财务接口
│   └── proxy.js      # 代理接口
├── components/       # 全局组件
│   ├── station-switch/  # 站点切换器
│   ├── status-tag/      # 状态标签
│   └── empty-state/     # 空状态
├── config/           # 配置文件
│   └── index.js      # 全局配置
├── pages/            # 页面
│   ├── tabbar/       # 底部导航页面
│   │   ├── home/     # 工作台
│   │   ├── cards/    # 卡密管理
│   │   └── profile/  # 我的
│   ├── card/         # 卡密相关（分包）
│   ├── finance/      # 财务相关（分包）
│   ├── proxy/        # 代理相关（分包）
│   └── common/       # 通用页面
│       └── login/    # 登录页
├── store/            # 状态管理
│   ├── user.js       # 用户状态
│   └── station.js    # 站点状态
├── styles/           # 全局样式
│   ├── variables.wxss  # 变量
│   └── common.wxss     # 通用样式
├── utils/            # 工具函数
│   └── format.js     # 格式化工具
├── app.js
├── app.json
└── app.wxss
```

## 快速开始

### 1. 安装依赖
```bash
npm install
```

### 2. 构建 npm
在微信开发者工具中：工具 -> 构建 npm

### 3. 配置后端 API
修改 `miniprogram/config/index.js` 中的 `API_BASE_URL`

### 4. 运行项目
在微信开发者工具中打开项目目录

## 待对接 API 列表

### 认证相关
- `POST /api/auth/login` - 登录
- `GET /api/auth/userinfo` - 获取用户信息

### 卡密相关
- `GET /api/cards` - 获取卡密列表
- `POST /api/cards/create` - 创建单个卡密
- `POST /api/cards/create/batch` - 批量创建卡密
- `GET /api/cards/:id` - 获取卡密详情
- `POST /api/cards/:id/destroy` - 销毁卡密
- `GET /api/cards/:id/logs` - 获取卡密操作日志

### 财务相关
- `POST /api/points/recharge` - 提交充值申请
- `GET /api/points/ledger` - 获取充值记录
- `GET /api/points/balance/:agentId` - 获取积分余额

### 代理相关
- `GET /api/proxy/list` - 获取下级代理列表
- `GET /api/audit/logs` - 获取操作日志

## 注意事项
1. 所有 API 请求会自动携带 `Authorization` 和 `X-Site-Id` 请求头
2. 站点切换后会自动更新 `X-Site-Id`，无需手动处理
3. Token 过期会自动跳转到登录页
4. 需要在微信公众平台配置服务器域名白名单

## 下一步开发
- [ ] 实现卡密创建页面（单个/批量）
- [ ] 实现卡密详情页面
- [ ] 实现充值申请页面
- [ ] 实现充值记录页面
- [ ] 实现下级代理列表页面
- [ ] 实现操作日志页面
- [ ] 对接后端 API
- [ ] 完善错误处理
- [ ] 添加加载状态优化
- [ ] 性能优化

## 联系方式
如有问题，请联系开发团队。
