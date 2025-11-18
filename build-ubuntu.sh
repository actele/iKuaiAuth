#!/bin/bash
# iKuai认证服务 - Ubuntu交叉编译脚本

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$PROJECT_DIR/dist"
APP_NAME="ikuai-auth"

echo "🔨 交叉编译 iKuai 认证服务 (Ubuntu x86_64)"
echo "=========================================="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go编译器"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 编译 Linux x86_64 版本
echo "🔧 编译 Linux x86_64 版本..."
cd "$PROJECT_DIR"

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

go build -ldflags="-s -w" -o "$OUTPUT_DIR/$APP_NAME" .

if [ $? -eq 0 ]; then
    echo "✅ 编译成功!"
    echo "📦 输出文件: $OUTPUT_DIR/$APP_NAME"
    echo "📏 文件大小: $(du -h "$OUTPUT_DIR/$APP_NAME" | cut -f1)"
    
    # 复制必要文件
    echo "📁 复制部署文件..."
    cp "$PROJECT_DIR/config.yaml" "$OUTPUT_DIR/"
    
    # 创建部署包
    echo "📦 创建部署包..."
    cd "$OUTPUT_DIR"
    tar -czf "${APP_NAME}-linux-amd64.tar.gz" "$APP_NAME" config.yaml
    
    echo ""
    echo "✅ 部署包创建成功: ${APP_NAME}-linux-amd64.tar.gz"
    echo ""
    echo "📋 部署说明:"
    echo "  1. 上传 dist/${APP_NAME}-linux-amd64.tar.gz 到Ubuntu服务器"
    echo "  2. 解压: tar -xzf ${APP_NAME}-linux-amd64.tar.gz"
    echo "  3. 运行安装脚本: sudo ./install-ubuntu.sh"
else
    echo "❌ 编译失败"
    exit 1
fi
