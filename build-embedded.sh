#!/bin/bash

# iKuai认证服务 - 嵌入式版本构建脚本
# 该脚本会将所有静态文件嵌入到Go二进制文件中

set -e

# 项目目录
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_DIR="$PROJECT_DIR/web"
OUTPUT_BINARY="$PROJECT_DIR/ikuai-auth-embedded"

echo "🔨 构建 iKuai 认证服务 (嵌入式版本)"
echo "================================"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go编译器"
    echo "请先安装Go: https://golang.org/dl/"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 检查项目结构
if [ ! -f "$PROJECT_DIR/main.go" ]; then
    echo "❌ 错误: 找不到main.go文件"
    exit 1
fi

# 确保web目录存在
if [ ! -d "$WEB_DIR" ]; then
    echo "❌ 错误: 找不到web目录"
    exit 1
fi

echo "✅ 项目结构检查完成"

# 检查必要文件
required_files=(
    "$WEB_DIR/index.html"
    "$WEB_DIR/auth.js"
    "$WEB_DIR/world-map.svg"
)

for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "❌ 错误: 找不到必要文件 $file"
        exit 1
    fi
done

echo "✅ 静态文件检查完成"

# 在项目根目录进行编译
cd "$PROJECT_DIR"

echo "🔧 编译中..."

# 设置编译参数
export CGO_ENABLED=0
export GOOS=$(go env GOOS)
export GOARCH=$(go env GOARCH)

# 编译
go build -ldflags="-s -w" -o "$OUTPUT_BINARY" .

if [ $? -eq 0 ]; then
    echo "✅ 编译成功!"
    echo ""
    echo "📦 输出文件: $OUTPUT_BINARY"
    echo "📏 文件大小: $(du -h "$OUTPUT_BINARY" | cut -f1)"
    echo ""
    echo "🚀 使用方法:"
    echo "   启动服务: ./start-embedded.sh start"
    echo "   查看状态: ./start-embedded.sh status"
    echo "   停止服务: ./start-embedded.sh stop"
    echo ""
    echo "💡 注意事项:"
    echo "   - 该可执行文件包含了所有静态资源"
    echo "   - 只需要config.yaml配置文件即可运行"
    echo "   - 无需额外的web目录"
else
    echo "❌ 编译失败"
    exit 1
fi