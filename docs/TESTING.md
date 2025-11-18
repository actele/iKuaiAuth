# 测试指南 🧪

本文档提供功能测试、故障排查和验证方法。

## 目录

- [快速测试](#快速测试)
- [功能测试](#功能测试)
- [API 测试](#api-测试)
- [Token 验证](#token-验证)
- [故障排查](#故障排查)
- [性能测试](#性能测试)

## 快速测试

### 1. 启动服务

```bash
# 方式 1: 直接运行
./ikuai-auth

# 方式 2: 使用 systemd（已安装）
sudo systemctl start ikuai-auth
```

### 2. 验证服务

```bash
# 检查服务状态
sudo systemctl status ikuai-auth

# 测试健康检查接口
curl http://localhost:8088/api/health

# 预期响应
{"status":"ok"}
```

### 3. 访问认证页面

```bash
# 在浏览器中访问
http://localhost:8088/

# 或使用 curl
curl http://localhost:8088/
```

## 功能测试

### 本地认证测试（Simple Mode）

#### 1. 配置测试用户

编辑 `config.yaml`：

```yaml
auth:
  method: "simple"
  simple_users:
    testuser: "testpass123"
    admin: "admin123"
```

#### 2. 重启服务

```bash
sudo systemctl restart ikuai-auth
```

#### 3. 测试登录

访问 `http://localhost:8088/` 并使用测试账号登录。

### API 认证测试

#### 1. 配置 API 认证

```yaml
auth:
  method: "api"
  api:
    url: "https://your-api.com/auth"
    method: "POST"
    timeout: 5
    headers:
      Content-Type: "application/json"
    body_template: |
      {
        "username": "{{username}}",
        "password": "{{password}}"
      }
    response:
      success_field: "success"
      success_value: true
      message_field: "message"
```

#### 2. 启用调试模式

```yaml
server:
  debug: true
```

#### 3. 查看日志

```bash
sudo journalctl -u ikuai-auth -f
```

日志应显示：

```
调用外部API认证: POST https://your-api.com/auth (用户: testuser)
API认证成功 - 用户: testuser
```

## API 测试

### 认证接口测试

```bash
curl -X POST http://localhost:8088/api/auth \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "user_ip": "192.168.1.100",
    "mac": "00:50:56:c0:00:08",
    "upload": 0,
    "download": 0,
    "timeout": 0,
    "release_type": 1
  }'
```

**成功响应示例：**

```json
{
  "success": true,
  "message": "认证成功",
  "data": {
    "release_url": "https://portal.ikuai8-wifi.com/Action/webauth-up?...",
    "token": "abc123def456...",
    "debug": false
  }
}
```

**失败响应示例：**

```json
{
  "success": false,
  "message": "用户名或密码错误"
}
```

### 健康检查测试

```bash
curl http://localhost:8088/api/health
```

**响应：**

```json
{
  "status": "ok"
}
```

## Token 验证

### 手动计算 Token

创建测试脚本 `test_token.go`：

```go
package main

import (
    "crypto/md5"
    "fmt"
    "sort"
    "strings"
)

func main() {
    params := map[string]string{
        "type":         "20",
        "user_id":      "testuser",
        "custom_name":  "testuser",
        "user_ip":      "192.168.1.100",
        "timestamp":    "1704441600",
        "mac":          "00:50:56:c0:00:08",
        "timeout":      "0",
        "upload":       "0",
        "download":     "0",
        "release_type": "1",
        "key":          "your_app_key_here",
    }
    
    // 排序参数
    var keys []string
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    // 构建查询字符串
    var parts []string
    for _, k := range keys {
        parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
    }
    query := strings.Join(parts, "&")
    
    // 计算 MD5
    hash := md5.Sum([]byte(query))
    token := fmt.Sprintf("%x", hash)
    
    fmt.Printf("Query String:\n%s\n\n", query)
    fmt.Printf("Token: %s\n", token)
}
```

运行测试：

```bash
go run test_token.go
```

### 验证 Token 一致性

1. 使用 `/api/auth` 接口获取 token
2. 使用上述脚本手动计算 token
3. 比较两个 token 是否一致

## 故障排查

### 服务无法启动

#### 症状

```bash
sudo systemctl status ikuai-auth
# 显示 failed 状态
```

#### 排查步骤

```bash
# 1. 查看错误日志
sudo journalctl -u ikuai-auth -n 50

# 2. 检查配置文件
cat /opt/ikuai-auth/config.yaml

# 3. 验证 YAML 语法
python3 -c "import yaml; yaml.safe_load(open('/opt/ikuai-auth/config.yaml'))"

# 4. 检查文件权限
ls -la /opt/ikuai-auth

# 5. 手动运行测试
sudo -u ikuai /opt/ikuai-auth/ikuai-auth
```

### 端口被占用

#### 症状

```
bind: address already in use
```

#### 解决方案

```bash
# 查找占用进程
sudo lsof -i :8088

# 或
sudo netstat -tlnp | grep 8088

# 杀死进程
sudo kill -9 <PID>

# 或修改配置文件中的端口
nano /opt/ikuai-auth/config.yaml
```

### 认证失败

#### 症状

登录时提示"用户名或密码错误"

#### 排查步骤

**Simple 模式：**

```bash
# 1. 确认配置文件中有该用户
grep -A 5 "simple_users" /opt/ikuai-auth/config.yaml

# 2. 检查密码是否匹配（区分大小写）

# 3. 重启服务
sudo systemctl restart ikuai-auth
```

**API 模式：**

```bash
# 1. 查看 API 调用日志
sudo journalctl -u ikuai-auth -f

# 2. 测试 API 连通性
curl -v https://your-api.com/auth

# 3. 检查 API 响应格式
# 确认 success_field 和 success_value 配置正确

# 4. 验证请求体模板
# 确认模板变量替换正确
```

### Token 错误

#### 症状

iKuai 路由器返回"Token 验证失败"

#### 排查步骤

```bash
# 1. 启用调试模式
nano /opt/ikuai-auth/config.yaml
# 设置 debug: true

# 2. 重启服务
sudo systemctl restart ikuai-auth

# 3. 查看生成的 Token
sudo journalctl -u ikuai-auth -f

# 4. 验证 AppKey 是否正确
# 确认 config.yaml 中的 app_key 与爱快云平台一致

# 5. 检查时间同步
timedatectl status
# 时间不同步可能导致 Token 验证失败
```

### 无法访问服务

#### 症状

浏览器无法打开认证页面

#### 排查步骤

```bash
# 1. 检查服务是否运行
sudo systemctl status ikuai-auth

# 2. 检查端口监听
sudo netstat -tlnp | grep 8088

# 3. 测试本地访问
curl http://localhost:8088/

# 4. 检查防火墙
sudo ufw status
# 或
sudo firewall-cmd --list-all

# 5. 开放端口（如需要）
sudo ufw allow 8088/tcp
# 或
sudo firewall-cmd --permanent --add-port=8088/tcp
sudo firewall-cmd --reload

# 6. 检查 Nginx 配置（如使用反向代理）
sudo nginx -t
```

## 性能测试

### 并发测试

使用 `ab` (Apache Bench) 进行压力测试：

```bash
# 安装 ab
sudo apt install apache2-utils

# 测试认证接口
ab -n 1000 -c 10 -p auth.json -T application/json \
  http://localhost:8088/api/auth

# auth.json 内容
echo '{
  "username": "testuser",
  "password": "testpass123",
  "user_ip": "192.168.1.100",
  "mac": "00:50:56:c0:00:08",
  "upload": 0,
  "download": 0,
  "timeout": 0,
  "release_type": 1
}' > auth.json
```

### 资源监控

```bash
# 查看服务资源占用
ps aux | grep ikuai-auth

# 实时监控
top -p $(pgrep ikuai-auth)

# 详细统计
systemctl status ikuai-auth
```

## 集成测试

### 完整认证流程测试

```bash
#!/bin/bash
# test_full_flow.sh

BASE_URL="http://localhost:8088"

echo "1. 测试健康检查..."
curl -s $BASE_URL/api/health | jq .

echo -e "\n2. 测试认证接口..."
RESPONSE=$(curl -s -X POST $BASE_URL/api/auth \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "user_ip": "192.168.1.100",
    "mac": "00:50:56:c0:00:08",
    "upload": 0,
    "download": 0,
    "timeout": 0,
    "release_type": 1
  }')

echo $RESPONSE | jq .

# 提取 release_url
RELEASE_URL=$(echo $RESPONSE | jq -r '.data.release_url')

echo -e "\n3. 放行 URL:"
echo $RELEASE_URL

# 可选：测试放行接口
# curl -v "$RELEASE_URL"
```

运行测试：

```bash
chmod +x test_full_flow.sh
./test_full_flow.sh
```

## 调试技巧

### 启用详细日志

```yaml
server:
  debug: true
```

### 实时日志监控

```bash
# 跟踪所有日志
sudo journalctl -u ikuai-auth -f

# 只看错误
sudo journalctl -u ikuai-auth -p err -f

# 查看最近 100 条
sudo journalctl -u ikuai-auth -n 100
```

### 浏览器开发者工具

1. 打开浏览器开发者工具（F12）
2. **Network 标签** - 查看 HTTP 请求
3. **Console 标签** - 查看 JavaScript 错误
4. **Application 标签** - 查看 LocalStorage（暗黑模式）

### 抓包分析

```bash
# 使用 tcpdump 抓包
sudo tcpdump -i any -A 'port 8088'

# 只看 POST 请求
sudo tcpdump -i any -A 'port 8088 and tcp[((tcp[12:1] & 0xf0) >> 2):4] = 0x504f5354'
```

## 自动化测试

### 创建测试脚本

```bash
#!/bin/bash
# automated_test.sh

set -e  # 遇到错误立即退出

echo "=== iKuai 认证服务自动化测试 ==="

# 1. 检查服务状态
echo -n "检查服务状态... "
if systemctl is-active --quiet ikuai-auth; then
    echo "✓ 运行中"
else
    echo "✗ 服务未运行"
    exit 1
fi

# 2. 测试健康检查
echo -n "测试健康检查... "
HEALTH=$(curl -s http://localhost:8088/api/health)
if echo $HEALTH | grep -q "ok"; then
    echo "✓ 通过"
else
    echo "✗ 失败"
    exit 1
fi

# 3. 测试认证
echo -n "测试认证接口... "
AUTH=$(curl -s -X POST http://localhost:8088/api/auth \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "user_ip": "192.168.1.100",
    "mac": "00:50:56:c0:00:08",
    "upload": 0,
    "download": 0,
    "timeout": 0,
    "release_type": 1
  }')

if echo $AUTH | grep -q "success"; then
    echo "✓ 通过"
else
    echo "✗ 失败"
    echo $AUTH | jq .
    exit 1
fi

echo -e "\n所有测试通过！✓"
```

### 定时测试

添加 cron 任务：

```bash
# 每小时运行一次测试
0 * * * * /opt/ikuai-auth/automated_test.sh >> /var/log/ikuai-test.log 2>&1
```

## 常见问题 FAQ

### Q: 如何测试 API 认证但没有外部 API？

A: 可以使用 [RequestBin](https://requestbin.com/) 或本地 mock 服务：

```python
# mock_api.py
from flask import Flask, request, jsonify
app = Flask(__name__)

@app.route('/auth', methods=['POST'])
def auth():
    data = request.json
    # 简单验证：用户名等于密码
    if data['username'] == data['password']:
        return jsonify({'success': True, 'message': 'ok'})
    return jsonify({'success': False, 'message': 'failed'})

app.run(port=5000)
```

### Q: 如何模拟 iKuai 路由器环境？

A: 手动访问认证页面时添加参数：

```
http://localhost:8088/?user_ip=192.168.1.100&mac=00:11:22:33:44:55&...
```

### Q: 如何验证 Token 是否正确？

A: 使用 Token 验证脚本（见上方）或启用调试模式查看详细输出。

## 支持

测试遇到问题？

- [提交 Issue](https://github.com/actele/iKuaiAuth/issues)
- [查看文档](../README.md)
- [架构说明](ARCHITECTURE.md)

---

**提示**: 生产环境测试前，请先在测试环境验证。
