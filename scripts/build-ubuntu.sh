#!/bin/bash
# iKuai认证服务 - Ubuntu交叉编译和打包脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$PROJECT_DIR/dist"
APP_NAME="ikuai-auth"
VERSION=$(date +"%Y%m%d_%H%M%S")

echo "🔨 编译和打包 iKuai 认证服务 (Ubuntu x86_64)"
echo "=========================================="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go编译器"
    exit 1
fi

echo "✅ Go 版本: $(go version)"
echo "📦 版本号: $VERSION"

# 创建输出目录
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR/$APP_NAME"

# 编译 Linux x86_64 版本
echo "🔧 编译 Linux x86_64 版本..."

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

cd "$PROJECT_DIR"
go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT_DIR/$APP_NAME/$APP_NAME" .

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功!"
echo "📏 文件大小: $(du -h "$OUTPUT_DIR/$APP_NAME/$APP_NAME" | cut -f1)"

# 复制必要文件到部署包
echo "📁 准备部署文件..."
cp "$PROJECT_DIR/config.yaml.example" "$OUTPUT_DIR/$APP_NAME/"
cp "$PROJECT_DIR/LICENSE" "$OUTPUT_DIR/$APP_NAME/" 2>/dev/null || true

# 创建安装脚本
cat > "$OUTPUT_DIR/$APP_NAME/install.sh" << 'EOF'
#!/bin/bash
set -e

APP_NAME="ikuai-auth"
INSTALL_DIR="/opt/$APP_NAME"
SERVICE_NAME="ikuai-auth"
SERVICE_USER="ikuai"

if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用 sudo 运行此脚本"
    exit 1
fi

echo "🚀 安装 iKuai 认证服务"
echo "========================"

# 创建服务用户
if ! id "$SERVICE_USER" &>/dev/null; then
    echo "👤 创建服务用户: $SERVICE_USER"
    useradd -r -s /bin/false $SERVICE_USER
fi

# 创建安装目录
echo "📁 创建目录: $INSTALL_DIR"
mkdir -p $INSTALL_DIR/logs

# 复制文件
echo "📦 安装程序文件..."
cp $APP_NAME $INSTALL_DIR/
chmod +x $INSTALL_DIR/$APP_NAME

# 配置文件处理
if [ -f "$INSTALL_DIR/config.yaml" ]; then
    echo "⚠️  保留现有配置文件"
    cp $INSTALL_DIR/config.yaml $INSTALL_DIR/config.yaml.backup
else
    echo "📝 安装配置文件"
    cp config.yaml.example $INSTALL_DIR/config.yaml
fi

# 设置权限
chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
chmod 600 $INSTALL_DIR/config.yaml

# 创建 systemd 服务
echo "⚙️  配置 systemd 服务..."
cat > /etc/systemd/system/$SERVICE_NAME.service << SERVICEEOF
[Unit]
Description=iKuai Authentication Service
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$APP_NAME
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
SERVICEEOF

# 重载并启动服务
systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl restart $SERVICE_NAME

echo ""
echo "✅ 安装完成!"
echo ""
echo "📋 后续步骤:"
echo "  1. 编辑配置: sudo nano $INSTALL_DIR/config.yaml"
echo "  2. 重启服务: sudo systemctl restart $SERVICE_NAME"
echo "  3. 查看状态: sudo systemctl status $SERVICE_NAME"
echo "  4. 查看日志: sudo journalctl -u $SERVICE_NAME -f"
echo ""
systemctl status $SERVICE_NAME --no-pager
EOF

chmod +x "$OUTPUT_DIR/$APP_NAME/install.sh"

# 创建更新脚本
cat > "$OUTPUT_DIR/$APP_NAME/update.sh" << 'EOF'
#!/bin/bash
set -e

APP_NAME="ikuai-auth"
INSTALL_DIR="/opt/$APP_NAME"
SERVICE_NAME="ikuai-auth"

if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用 sudo 运行此脚本"
    exit 1
fi

if [ ! -d "$INSTALL_DIR" ]; then
    echo "❌ 未检测到已安装的服务，请先运行 install.sh"
    exit 1
fi

echo "🔄 更新 iKuai 认证服务"
echo "========================"

# 停止服务
echo "⏸️  停止服务..."
systemctl stop $SERVICE_NAME

# 备份旧版本
echo "💾 备份旧版本..."
cp $INSTALL_DIR/$APP_NAME $INSTALL_DIR/${APP_NAME}.backup

# 更新程序
echo "📦 更新程序文件..."
cp $APP_NAME $INSTALL_DIR/
chmod +x $INSTALL_DIR/$APP_NAME
chown ikuai:ikuai $INSTALL_DIR/$APP_NAME

# 启动服务
echo "▶️  启动服务..."
systemctl start $SERVICE_NAME

echo ""
echo "✅ 更新完成!"
echo ""
systemctl status $SERVICE_NAME --no-pager
EOF

chmod +x "$OUTPUT_DIR/$APP_NAME/update.sh"

# 创建卸载脚本
cat > "$OUTPUT_DIR/$APP_NAME/uninstall.sh" << 'EOF'
#!/bin/bash
set -e

APP_NAME="ikuai-auth"
INSTALL_DIR="/opt/$APP_NAME"
SERVICE_NAME="ikuai-auth"

if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用 sudo 运行此脚本"
    exit 1
fi

echo "🗑️  卸载 iKuai 认证服务"
echo "========================"

read -p "确认卸载? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 0
fi

# 停止并禁用服务
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "⏸️  停止服务..."
    systemctl stop $SERVICE_NAME
fi

if systemctl is-enabled --quiet $SERVICE_NAME; then
    echo "🔧 禁用服务..."
    systemctl disable $SERVICE_NAME
fi

# 删除服务文件
echo "🗑️  删除服务文件..."
rm -f /etc/systemd/system/$SERVICE_NAME.service
systemctl daemon-reload

# 删除安装目录
echo "🗑️  删除程序文件..."
rm -rf $INSTALL_DIR

echo ""
echo "✅ 卸载完成!"
EOF

chmod +x "$OUTPUT_DIR/$APP_NAME/uninstall.sh"

# 创建部署包
echo "📦 创建部署包..."
cd "$OUTPUT_DIR"
tar -czf "${APP_NAME}-linux-amd64.tar.gz" "$APP_NAME"

PACKAGE_SIZE=$(du -h "${APP_NAME}-linux-amd64.tar.gz" | cut -f1)

echo ""
echo "✅ 打包完成!"
echo "=========================================="
echo "📦 部署包: dist/${APP_NAME}-linux-amd64.tar.gz"
echo "📏 大小: $PACKAGE_SIZE"
echo "📅 版本: $VERSION"
echo ""
echo "📋 使用说明:"
echo "  1. 上传到服务器: scp dist/${APP_NAME}-linux-amd64.tar.gz user@server:/tmp/"
echo "  2. 解压: tar -xzf ${APP_NAME}-linux-amd64.tar.gz"
echo "  3. 安装: cd $APP_NAME && sudo ./install.sh"
echo "  4. 更新: cd $APP_NAME && sudo ./update.sh"
echo ""
