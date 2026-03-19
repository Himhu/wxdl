#!/bin/bash

echo "================================"
echo "代理商管理系统 - 快速启动脚本"
echo "================================"
echo ""

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未安装 Go，请先安装 Go 1.21+"
    exit 1
fi

echo "✅ Go 版本: $(go version)"
echo ""

# 检查 MySQL
if ! command -v mysql &> /dev/null; then
    echo "⚠️  警告: 未检测到 MySQL，请确保 MySQL 已安装并运行"
fi

# 创建 .env 文件
if [ ! -f .env ]; then
    echo "📝 创建 .env 配置文件..."
    cp .env.example .env
    echo "✅ 已创建 .env 文件，请编辑配置后再运行"
    echo ""
    echo "需要配置的项目："
    echo "  - DB_PASSWORD: 数据库密码"
    echo "  - JWT_SECRET: JWT 密钥"
    echo "  - WECHAT_APPID: 微信小程序 AppID"
    echo "  - WECHAT_SECRET: 微信小程序 Secret"
    echo ""
    exit 0
fi

# 安装依赖
echo "📦 安装 Go 依赖..."
go mod download
if [ $? -ne 0 ]; then
    echo "❌ 依赖安装失败"
    exit 1
fi
echo "✅ 依赖安装完成"
echo ""

# 创建日志目录
mkdir -p logs

# 提示初始化数据库
echo "📊 数据库初始化"
echo "请手动执行以下命令初始化数据库："
echo ""
echo "  mysql -u root -p -e \"CREATE DATABASE agent_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;\""
echo "  mysql -u root -p agent_system < scripts/init.sql"
echo ""
read -p "数据库已初始化？(y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "请先初始化数据库后再运行"
    exit 0
fi

# 启动服务
echo ""
echo "🚀 启动服务..."
echo "================================"
echo ""

go run cmd/server/main.go
