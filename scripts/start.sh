#!/bin/bash

# KeyDesk 快速启动脚本

set -e

echo "🚀 启动 KeyDesk 激活码管理系统..."

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.21 或更高版本"
    exit 1
fi

# 检查PostgreSQL是否运行
if ! pg_isready -h localhost -p 5432 &> /dev/null; then
    echo "⚠️  PostgreSQL 未运行，尝试启动..."
    
    # 尝试使用Docker启动PostgreSQL
    if command -v docker &> /dev/null; then
        echo "🐳 使用 Docker 启动 PostgreSQL..."
        docker run -d \
            --name keydesk-postgres \
            -e POSTGRES_DB=keydesk \
            -e POSTGRES_USER=postgres \
            -e POSTGRES_PASSWORD=password \
            -p 5432:5432 \
            postgres:15-alpine
        
        echo "⏳ 等待 PostgreSQL 启动..."
        sleep 10
    else
        echo "❌ 请确保 PostgreSQL 正在运行在 localhost:5432"
        exit 1
    fi
fi

# 安装依赖
echo "📦 安装依赖..."
go mod tidy

# 创建配置文件
if [ ! -f "config/config.yaml" ]; then
    echo "📝 创建配置文件..."
    cp config/config.example.yaml config/config.yaml
    echo "✅ 配置文件已创建，请根据需要修改 config/config.yaml"
fi

# 构建应用
echo "🔨 构建应用..."
go build -o build/keydesk cmd/server/main.go

# 启动应用
echo "🌟 启动 KeyDesk 服务器..."
echo "📱 访问地址: http://localhost:8080"
echo "🔧 API 文档: http://localhost:8080/api/v1/health"
echo ""
echo "💡 默认管理员账户:"
echo "   邮箱: admin@keydesk.com"
echo "   密码: admin123"
echo ""
echo "按 Ctrl+C 停止服务器"

./build/keydesk 