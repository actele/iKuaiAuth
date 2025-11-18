# 部署指南 📦

本文档提供 iKuai 认证服务在不同环境下的部署方法。

## 目录

- [Ubuntu/Debian 部署](#ubuntudebian-部署)
- [CentOS/RHEL 部署](#centosrhel-部署)
- [Docker 部署](#docker-部署)
- [Nginx 反向代理](#nginx-反向代理)
- [服务管理](#服务管理)
- [故障排查](#故障排查)

## Ubuntu/Debian 部署

### 快速部署

#### 1. 下载部署包

```bash
# 下载最新版本
wget https://github.com/actele/iKuaiAuth/releases/latest/download/ikuai-auth-linux-amd64.tar.gz

# 解压
tar -xzf ikuai-auth-linux-amd64.tar.gz
cd ikuai-auth-linux-amd64
```

#### 2. 配置服务

```bash
# 复制配置模板
cp config.yaml.example config.yaml

# 编辑配置（重要）
nano config.yaml
```

**必须修改的配置项：**

```yaml
server:
  host: "0.0.0.0"
  port: "8088"    # 可以改为其他端口
  debug: false    # 生产环境设为 false

ikuai:
  app_key: "your_app_key_here"  # 从爱快云平台获取

auth:
  method: "simple"  # 或 "api"
  simple_users:
    admin: "strong_password_here"  # 修改为强密码
```

#### 3. 自动安装

```bash
chmod +x install-ubuntu.sh
sudo ./install-ubuntu.sh
```

**安装脚本会自动完成：**
- ✅ 创建服务用户 `ikuai`
- ✅ 安装到 `/opt/ikuai-auth`
- ✅ 创建 systemd 服务
- ✅ 启动服务并设置开机自启
- ✅ 配置日志目录

#### 4. 验证安装

```bash
# 查看服务状态
sudo systemctl status ikuai-auth

# 查看日志
sudo journalctl -u ikuai-auth -n 20

# 测试访问
curl http://localhost:8088/api/health
```

### 手动部署

如果不想使用安装脚本，可以手动部署：

```bash
# 1. 创建用户
sudo useradd -r -s /bin/false ikuai

# 2. 创建目录
sudo mkdir -p /opt/ikuai-auth
sudo mkdir -p /opt/ikuai-auth/logs

# 3. 复制文件
sudo cp ikuai-auth /opt/ikuai-auth/
sudo cp config.yaml /opt/ikuai-auth/

# 4. 设置权限
sudo chown -R ikuai:ikuai /opt/ikuai-auth
sudo chmod +x /opt/ikuai-auth/ikuai-auth

# 5. 创建 systemd 服务
sudo tee /etc/systemd/system/ikuai-auth.service > /dev/null <<'EOF'
[Unit]
Description=iKuai Authentication Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=ikuai
Group=ikuai
WorkingDirectory=/opt/ikuai-auth
ExecStart=/opt/ikuai-auth/ikuai-auth
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ikuai-auth

[Install]
WantedBy=multi-user.target
EOF

# 6. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable ikuai-auth
sudo systemctl start ikuai-auth
```

## CentOS/RHEL 部署

基本步骤与 Ubuntu 相同，主要区别：

```bash
# 防火墙配置
sudo firewall-cmd --permanent --add-port=8088/tcp
sudo firewall-cmd --reload

# SELinux 配置（如果启用）
sudo semanage port -a -t http_port_t -p tcp 8088
```

## Docker 部署

### 使用 Docker Compose（推荐）

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  ikuai-auth:
    image: ghcr.io/actele/ikuai-auth:latest
    container_name: ikuai-auth
    restart: unless-stopped
    ports:
      - "8088:8088"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ./logs:/app/logs
    environment:
      - TZ=Asia/Shanghai
```

启动服务：

```bash
docker-compose up -d
```

### 手动 Docker 部署

```bash
# 拉取镜像
docker pull ghcr.io/actele/ikuai-auth:latest

# 运行容器
docker run -d \
  --name ikuai-auth \
  --restart unless-stopped \
  -p 8088:8088 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -v $(pwd)/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  ghcr.io/actele/ikuai-auth:latest
```

### 自定义构建

```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ikuai-auth .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/ikuai-auth .
COPY --from=builder /build/config.yaml.example .

EXPOSE 8088
CMD ["./ikuai-auth"]
```

## Nginx 反向代理

### 基础配置

创建 `/etc/nginx/sites-available/ikuai-auth`：

```nginx
server {
    listen 80;
    server_name auth.yourdomain.com;

    # 访问日志
    access_log /var/log/nginx/ikuai-auth.access.log;
    error_log /var/log/nginx/ikuai-auth.error.log;

    location / {
        proxy_pass http://127.0.0.1:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

启用配置：

```bash
sudo ln -s /etc/nginx/sites-available/ikuai-auth /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### HTTPS 配置（Let's Encrypt）

```bash
# 安装 Certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d auth.yourdomain.com

# 自动续期
sudo systemctl enable certbot.timer
```

Nginx 配置会自动更新为：

```nginx
server {
    listen 443 ssl http2;
    server_name auth.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/auth.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/auth.yourdomain.com/privkey.pem;
    
    # SSL 配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    location / {
        proxy_pass http://127.0.0.1:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
    }
}

server {
    listen 80;
    server_name auth.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

## 服务管理

### 基本命令

```bash
# 启动服务
sudo systemctl start ikuai-auth

# 停止服务
sudo systemctl stop ikuai-auth

# 重启服务
sudo systemctl restart ikuai-auth

# 查看状态
sudo systemctl status ikuai-auth

# 开机自启
sudo systemctl enable ikuai-auth

# 禁用自启
sudo systemctl disable ikuai-auth
```

### 日志管理

```bash
# 查看实时日志
sudo journalctl -u ikuai-auth -f

# 查看最近 50 条
sudo journalctl -u ikuai-auth -n 50

# 查看特定时间
sudo journalctl -u ikuai-auth --since "2025-01-01 10:00:00"

# 查看今天的日志
sudo journalctl -u ikuai-auth --since today
```

### 配置更新

```bash
# 1. 编辑配置
sudo nano /opt/ikuai-auth/config.yaml

# 2. 重启服务
sudo systemctl restart ikuai-auth

# 3. 验证
sudo journalctl -u ikuai-auth -n 20
```

### 日志轮转

创建 `/etc/logrotate.d/ikuai-auth`：

```
/opt/ikuai-auth/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 ikuai ikuai
    sharedscripts
    postrotate
        systemctl reload ikuai-auth > /dev/null 2>&1 || true
    endscript
}
```

## 防火墙配置

### UFW (Ubuntu/Debian)

```bash
# 开放端口
sudo ufw allow 8088/tcp

# 限制来源 IP
sudo ufw allow from 192.168.1.0/24 to any port 8088 proto tcp

# 查看规则
sudo ufw status
```

### Firewalld (CentOS/RHEL)

```bash
# 开放端口
sudo firewall-cmd --permanent --add-port=8088/tcp
sudo firewall-cmd --reload

# 限制来源 IP
sudo firewall-cmd --permanent --add-rich-rule='
  rule family="ipv4"
  source address="192.168.1.0/24"
  port protocol="tcp" port="8088" accept'
sudo firewall-cmd --reload
```

## 故障排查

### 服务无法启动

```bash
# 1. 查看错误日志
sudo journalctl -u ikuai-auth -n 50 --no-pager

# 2. 检查配置文件
cat /opt/ikuai-auth/config.yaml

# 3. 检查文件权限
ls -la /opt/ikuai-auth

# 4. 手动运行测试
sudo -u ikuai /opt/ikuai-auth/ikuai-auth
```

### 端口被占用

```bash
# 查看占用进程
sudo netstat -tlnp | grep 8088
# 或
sudo lsof -i :8088

# 杀死进程
sudo kill -9 <PID>
```

### 配置错误

```bash
# 检查 YAML 语法
python3 -c "import yaml; yaml.safe_load(open('/opt/ikuai-auth/config.yaml'))"
```

### 权限问题

```bash
# 修复权限
sudo chown -R ikuai:ikuai /opt/ikuai-auth
sudo chmod +x /opt/ikuai-auth/ikuai-auth
sudo chmod 644 /opt/ikuai-auth/config.yaml
```

### 连接问题

```bash
# 测试本地访问
curl http://localhost:8088/api/health

# 测试外部访问
curl http://server-ip:8088/api/health

# 检查防火墙
sudo iptables -L -n | grep 8088
```

## 卸载

使用卸载脚本：

```bash
cd /opt/ikuai-auth
sudo ./uninstall-ubuntu.sh
```

手动卸载：

```bash
# 停止并禁用服务
sudo systemctl stop ikuai-auth
sudo systemctl disable ikuai-auth

# 删除服务文件
sudo rm /etc/systemd/system/ikuai-auth.service
sudo systemctl daemon-reload

# 删除安装目录
sudo rm -rf /opt/ikuai-auth

# 删除用户（可选）
sudo userdel ikuai
```

## 性能优化

### 系统参数调优

编辑 `/etc/sysctl.conf`：

```bash
# 增加文件描述符限制
fs.file-max = 65535

# TCP 优化
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30
net.core.somaxconn = 1024
```

应用配置：

```bash
sudo sysctl -p
```

### 进程限制

编辑 `/etc/security/limits.conf`：

```
ikuai soft nofile 65535
ikuai hard nofile 65535
```

## 监控和告警

### 服务监控脚本

创建 `/usr/local/bin/check-ikuai-auth.sh`：

```bash
#!/bin/bash

if ! systemctl is-active --quiet ikuai-auth; then
    echo "iKuai Auth service is down! Restarting..."
    systemctl restart ikuai-auth
    
    # 发送告警（可选）
    # mail -s "iKuai Auth Alert" admin@example.com <<< "Service restarted"
fi
```

添加 cron 任务：

```bash
# 每 5 分钟检查一次
*/5 * * * * /usr/local/bin/check-ikuai-auth.sh
```

## 安全建议

1. **修改默认密码** - 不要使用示例密码
2. **限制访问源** - 使用防火墙规则限制访问
3. **启用 HTTPS** - 使用 Nginx + Let's Encrypt
4. **定期更新** - 关注新版本发布
5. **日志审计** - 定期检查异常访问
6. **备份配置** - 定期备份 config.yaml

## 支持

如遇问题，请访问：
- [GitHub Issues](https://github.com/actele/iKuaiAuth/issues)
- [项目文档](../README.md)
