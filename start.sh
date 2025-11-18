#!/bin/bash

# iKuai 认证服务 - 本地开发启动脚本

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="config.yaml"
BINARY_NAME="ikuai-auth"

echo "🚀 启动 iKuai 认证服务 (开发模式)"
echo "======================================"

# 检查配置文件
if [ ! -f "$PROJECT_DIR/$CONFIG_FILE" ]; then
    echo "❌ 错误: 找不到配置文件 $CONFIG_FILE"
    echo "💡 请创建 config.yaml 配置文件"
    exit 1
fi

echo "✅ 配置文件: $CONFIG_FILE"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go编译器"
    echo "请先安装Go: https://golang.org/dl/"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 编译并运行
echo "🔧 编译服务..."
cd "$PROJECT_DIR"

# 编译
go build -o "$BINARY_NAME" . || {
    echo "❌ 编译失败"
    exit 1
}

echo "✅ 编译成功"
echo ""
echo "🌐 启动服务..."
echo "======================================"

# 运行服务
./$BINARY_NAME
