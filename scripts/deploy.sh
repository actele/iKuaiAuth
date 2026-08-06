#!/bin/bash
# iKuai认证服务 - 快速部署脚本
# 自动编译、打包、上传并部署到服务器
# 用法: ./deploy.sh user@host

set -e

# 配置变量（必须通过命令行参数或环境变量传入，避免硬编码内网地址泄露）
#   用法:   ./deploy.sh user@host
#   或:     IKUAI_SERVER=user@host ./deploy.sh
SERVER_PATH="/mnt/ikuai-auth"
REMOTE_SERVICE_PATH="/opt/ikuai-auth"

# 解析 server 地址（命令行参数优先，其次环境变量）
if [ -n "$1" ]; then
    SERVER_TARGET="$1"
elif [ -n "$IKUAI_SERVER" ]; then
    SERVER_TARGET="$IKUAI_SERVER"
else
    echo "❌ 缺少服务器地址"
    echo "用法: $0 user@host"
    echo "或者: IKUAI_SERVER=user@host $0"
    exit 1
fi

# 解析 user@host 格式
SERVER_USER="${SERVER_TARGET%%@*}"
SERVER_HOST="${SERVER_TARGET##*@}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "🚀 iKuai 认证服务 - 快速部署"
echo "=========================================="
echo "📍 服务器: $SERVER_USER@$SERVER_HOST"
echo "📂 临时目录: $SERVER_PATH"
echo "📂 安装目录: $REMOTE_SERVICE_PATH"
echo ""

# 步骤1: 编译打包
echo "📦 步骤 1/4: 编译打包..."
cd "$PROJECT_DIR"
./scripts/build-ubuntu.sh

if [ ! -f "dist/ikuai-auth-linux-amd64.tar.gz" ]; then
    echo "❌ 打包文件不存在"
    exit 1
fi

echo "✅ 编译打包完成"
echo ""

# 步骤2: 上传到服务器
echo "📤 步骤 2/4: 上传到服务器..."
scp dist/ikuai-auth-linux-amd64.tar.gz $SERVER_USER@$SERVER_HOST:$SERVER_PATH/

if [ $? -ne 0 ]; then
    echo "❌ 上传失败"
    exit 1
fi

echo "✅ 上传完成"
echo ""

# 步骤3: 解压
echo "📂 步骤 3/4: 解压文件..."
ssh $SERVER_USER@$SERVER_HOST << 'ENDSSH'
cd /mnt/ikuai-auth
tar -xzf ikuai-auth-linux-amd64.tar.gz
echo "✅ 解压完成"
ENDSSH

echo ""

# 步骤4: 部署（安装或更新）
echo "🔄 步骤 4/4: 部署服务..."
ssh $SERVER_USER@$SERVER_HOST << 'ENDSSH'
set -e

cd /mnt/ikuai-auth/ikuai-auth

# 检查服务是否已安装
if [ -d "/opt/ikuai-auth" ]; then
    echo "🔄 检测到已安装服务，执行更新..."
    ./update.sh
else
    echo "📦 首次安装..."
    ./install.sh
fi

echo ""
echo "✅ 部署完成！"
echo ""
echo "📋 服务状态:"
systemctl status ikuai-auth --no-pager -l
ENDSSH

echo ""
echo "=========================================="
echo "🎉 部署成功完成！"
echo ""
echo "📋 快捷命令:"
echo "  查看日志: ssh $SERVER_USER@$SERVER_HOST 'journalctl -u ikuai-auth -f'"
echo "  重启服务: ssh $SERVER_USER@$SERVER_HOST 'systemctl restart ikuai-auth'"
echo "  查看状态: ssh $SERVER_USER@$SERVER_HOST 'systemctl status ikuai-auth'"
echo ""
