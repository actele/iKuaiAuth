# API 认证配置指南 🔐

本文档详细说明如何配置外部 API 认证，将 iKuai 认证系统与您现有的用户系统集成。

## 目录

- [概述](#概述)
- [配置方法](#配置方法)
- [API 要求](#api-要求)
- [配置示例](#配置示例)
- [测试验证](#测试验证)
- [故障排查](#故障排查)
- [示例实现](#示例实现)

## 概述

### 认证模式对比

| 特性 | Simple（本地认证） | API（外部认证） |
|------|-------------------|----------------|
| **配置难度** | ⭐ 简单 | ⭐⭐⭐ 中等 |
| **用户管理** | 配置文件 | 外部系统 |
| **适用场景** | 小型部署、测试 | 企业集成、统一认证 |
| **扩展性** | ❌ 低 | ✅ 高 |
| **依赖性** | ✅ 无外部依赖 | ❌ 依赖外部 API |

### 工作流程

```
用户登录 → iKuai 认证服务 → 调用外部 API → 验证用户 → 返回结果
```

## 配置方法

### 1. 修改配置文件

编辑 `config.yaml`，将认证方式改为 `api`：

```yaml
auth:
  method: "api"  # 从 "simple" 改为 "api"
  
  api:
    # API 端点配置
    url: "https://your-domain.com/api/auth/verify"
    method: "POST"  # GET 或 POST
    timeout: 5      # 超时时间（秒）
    
    # 自定义请求头
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer your-api-token-here"
      X-Custom-Header: "custom-value"
    
    # 请求体模板（仅 POST）
    body_template: |
      {
        "username": "{{username}}",
        "password": "{{password}}"
      }
    
    # 响应验证配置
    response:
      success_field: "success"    # 成功标识字段名
      success_value: true          # 成功时的值
      message_field: "message"     # 错误消息字段名（可选）
```

### 2. 配置参数说明

#### API 配置

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `url` | string | ✅ 是 | 认证 API 的完整 URL |
| `method` | string | ✅ 是 | HTTP 方法：`GET` 或 `POST` |
| `timeout` | int | ⚪ 否 | 超时时间（秒），默认 5 |
| `headers` | map | ⚪ 否 | 自定义 HTTP 请求头 |
| `body_template` | string | POST 时必填 | 请求体 JSON 模板 |

#### 响应配置

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `success_field` | string | ✅ 是 | 响应中表示成功的字段名 |
| `success_value` | any | ✅ 是 | 成功时该字段的值 |
| `message_field` | string | ⚪ 否 | 失败时获取错误信息的字段名 |

### 3. 模板变量

在 `body_template` 中可以使用以下变量：

| 变量 | 说明 | 示例 |
|------|------|------|
| `{{username}}` | 用户输入的用户名 | `admin` |
| `{{password}}` | 用户输入的密码 | `password123` |

**示例：**

```yaml
body_template: |
  {
    "user": "{{username}}",
    "pass": "{{password}}",
    "client": "ikuai-auth"
  }
```

实际发送时会替换为：

```json
{
  "user": "admin",
  "pass": "password123",
  "client": "ikuai-auth"
}
```

## API 要求

### 请求格式

#### GET 请求

如果使用 GET 方法，用户名和密码会作为 URL 参数传递：

```
GET https://api.example.com/auth?username=admin&password=pass123
```

配置示例：

```yaml
api:
  url: "https://api.example.com/auth"
  method: "GET"
  headers:
    Authorization: "Bearer token123"
```

#### POST 请求（推荐）

使用 POST 方法时，数据在请求体中发送：

```yaml
api:
  url: "https://api.example.com/auth/verify"
  method: "POST"
  headers:
    Content-Type: "application/json"
  body_template: |
    {
      "username": "{{username}}",
      "password": "{{password}}"
    }
```

### 响应格式

API 必须返回 JSON 格式，包含明确的成功标识字段。

#### 标准响应格式

**成功响应：**

```json
{
  "success": true,
  "message": "认证成功",
  "data": {
    "user_id": "12345",
    "username": "admin",
    "role": "admin"
  }
}
```

**失败响应：**

```json
{
  "success": false,
  "message": "用户名或密码错误"
}
```

#### 验证逻辑

系统会按以下步骤验证响应：

1. ✅ 检查 HTTP 状态码是否为 200
2. ✅ 解析 JSON 响应
3. ✅ 提取 `success_field` 指定的字段
4. ✅ 比较字段值与 `success_value` 是否匹配
5. ❌ 如果不匹配，从 `message_field` 获取错误信息

## 配置示例

### 示例 1: 标准 RESTful API

```yaml
auth:
  method: "api"
  api:
    url: "https://api.company.com/v1/auth/verify"
    method: "POST"
    timeout: 5
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer sk-1234567890abcdef"
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

### 示例 2: 使用状态码

如果 API 返回格式为：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

配置：

```yaml
auth:
  method: "api"
  api:
    url: "https://api.example.com/auth"
    method: "POST"
    body_template: |
      {
        "user": "{{username}}",
        "pwd": "{{password}}"
      }
    response:
      success_field: "code"
      success_value: 0
      message_field: "msg"
```

### 示例 3: 字符串状态

如果 API 返回格式为：

```json
{
  "status": "ok",
  "error": null
}
```

配置：

```yaml
auth:
  method: "api"
  api:
    url: "https://auth.service.com/check"
    method: "POST"
    body_template: |
      {
        "username": "{{username}}",
        "password": "{{password}}"
      }
    response:
      success_field: "status"
      success_value: "ok"
      message_field: "error"
```

### 示例 4: GET 请求

```yaml
auth:
  method: "api"
  api:
    url: "https://simple-auth.com/verify"
    method: "GET"
    timeout: 3
    headers:
      X-API-Key: "your-api-key"
    response:
      success_field: "authenticated"
      success_value: true
      message_field: "reason"
```

### 示例 5: 网络设备认证（Device Authentication）

如果您的 API 使用网络设备认证格式，返回用户详细信息：

```yaml
auth:
  method: "api"
  api:
    url: "http://192.168.1.100:8000/api/v1/network-auth/authenticate"
    method: "POST"
    timeout: 5
    headers:
      Content-Type: "application/json"
      X-API-Version: "v1"
    body_template: |
      {
        "device_number": "",
        "user_account": "{{username}}",
        "vpn_number": "",
        "mac_address": "{{password}}",
        "ip_address": "auto",
        "system_id": 1
      }
    response:
      success_field: "code"
      success_value: 0
      message_field: "msg"
```

**API 响应格式：**

```json
{
  "code": 0,
  "msg": "认证成功",
  "data": {
    "allowed": true,
    "reason": "",
    "user_id": "U1001",
    "username": "test_user",
    "real_name": "张三",
    "phone": "13800138000",
    "custom_name": "测试用户",
    "device_number": "1001",
    "assigned_to": "李四",
    "vpn_client": "OpenVPN"
  }
}
```

**用户信息映射：**

系统会自动提取 API 返回的用户信息并映射到 iKuai 认证参数：

| API 字段 | iKuai 参数 | 说明 |
|----------|-----------|------|
| `data.username` | `user_id` | 用户标识 |
| `data.username` | `custom_name` | 自定义名称 |
| `data.real_name` | `name` | 真实姓名 |
| `data.phone` | `phone` | 联系电话 |

如果 API 未返回用户信息，系统将使用默认格式 `1020004_用户名`。

## 测试验证

### 1. 启用调试模式

```yaml
server:
  debug: true
```

### 2. 查看日志

```bash
# 查看实时日志
sudo journalctl -u ikuai-auth -f

# 或查看最近日志
sudo journalctl -u ikuai-auth -n 50
```

调试模式下会输出：

```
调用外部API认证: POST https://api.example.com/verify (用户: admin)
API认证成功 - 用户: admin, IP: 192.168.1.100
```

### 3. 使用 curl 测试 API

在集成前，先用 curl 测试您的 API：

```bash
curl -X POST https://your-api.com/auth/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "username": "test",
    "password": "test123"
  }'
```

确认响应格式符合预期。

### 4. 测试认证流程

1. 访问认证页面: `http://your-server:8088/`
2. 输入测试用户名和密码
3. 查看日志确认 API 调用
4. 验证认证结果

## 故障排查

### 问题 1: API 请求超时

**症状：**

```
外部API认证失败 - 用户: admin, 错误: context deadline exceeded
```

**解决方案：**

1. 检查网络连接：
   ```bash
   curl -v https://your-api.com/endpoint
   ```

2. 增加超时时间：
   ```yaml
   api:
     timeout: 10  # 从 5 秒改为 10 秒
   ```

3. 确认 API 地址正确

### 问题 2: 响应格式错误

**症状：**

```
外部API认证失败 - 用户: admin, 错误: 响应格式错误
```

**解决方案：**

1. 启用调试模式查看实际响应
2. 确认 `success_field` 字段名正确
3. 检查 `success_value` 值的类型（布尔值、数字、字符串）

### 问题 3: 认证失败但应该成功

**症状：**

API 返回成功，但系统显示认证失败

**解决方案：**

1. 查看 API 实际返回内容
2. 确认 `success_value` 的类型和值：
   ```yaml
   # 布尔值
   success_value: true
   
   # 数字
   success_value: 0
   
   # 字符串
   success_value: "ok"
   ```

### 问题 4: 请求头未生效

**症状：**

API 返回 401 Unauthorized

**解决方案：**

检查 headers 配置格式：

```yaml
headers:
  Content-Type: "application/json"
  Authorization: "Bearer token123"  # 确保有引号
```

## 示例实现

### Python Flask

```python
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/api/auth/verify', methods=['POST'])
def verify():
    data = request.json
    username = data.get('username')
    password = data.get('password')
    
    # 您的认证逻辑
    if check_credentials(username, password):
        return jsonify({
            'success': True,
            'message': '认证成功',
            'data': {
                'user_id': username,
                'role': 'user'
            }
        })
    else:
        return jsonify({
            'success': False,
            'message': '用户名或密码错误'
        }), 200  # 注意：返回 200，错误由 success 字段标识

def check_credentials(username, password):
    # 实现您的认证逻辑
    # 例如：查询数据库、LDAP 等
    return username == 'admin' and password == 'admin123'

if __name__ == '__main__':
    app.run(port=5000)
```

### Node.js Express

```javascript
const express = require('express');
const app = express();

app.use(express.json());

app.post('/api/auth/verify', async (req, res) => {
    const { username, password } = req.body;
    
    try {
        // 您的认证逻辑
        const isValid = await checkCredentials(username, password);
        
        if (isValid) {
            res.json({
                success: true,
                message: '认证成功',
                data: { user_id: username }
            });
        } else {
            res.json({
                success: false,
                message: '用户名或密码错误'
            });
        }
    } catch (error) {
        res.json({
            success: false,
            message: '服务器错误'
        });
    }
});

async function checkCredentials(username, password) {
    // 实现您的认证逻辑
    return username === 'admin' && password === 'admin123';
}

app.listen(5000, () => {
    console.log('Auth API running on port 5000');
});
```

### PHP Laravel

```php
<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;

class AuthController extends Controller
{
    public function verify(Request $request)
    {
        $username = $request->input('username');
        $password = $request->input('password');
        
        // 您的认证逻辑
        if ($this->checkCredentials($username, $password)) {
            return response()->json([
                'success' => true,
                'message' => '认证成功',
                'data' => [
                    'user_id' => $username
                ]
            ]);
        } else {
            return response()->json([
                'success' => false,
                'message' => '用户名或密码错误'
            ]);
        }
    }
    
    private function checkCredentials($username, $password)
    {
        // 实现您的认证逻辑
        // 例如：Auth::attempt(['username' => $username, 'password' => $password])
        return $username === 'admin' && $password === 'admin123';
    }
}
```

## 安全建议

1. **使用 HTTPS**
   - API 地址必须使用 `https://` 协议
   - 避免明文传输密码

2. **添加认证**
   - 在 `headers` 中添加 API Key 或 Bearer Token
   - 防止未授权访问

3. **限流保护**
   - 在 API 端实现请求频率限制
   - 防止暴力破解

4. **日志审计**
   - 记录所有认证请求
   - 用于安全审计和问题排查

5. **密码加密**
   - 考虑在传输前对密码进行加密
   - 或使用更安全的认证方式（如 OAuth2）

## 切换回本地认证

如需切换回本地用户认证，修改配置：

```yaml
auth:
  method: "simple"
  simple_users:
    admin: "admin123"
    user01: "password456"
```

重启服务使配置生效：

```bash
sudo systemctl restart ikuai-auth
```

## 支持

遇到问题？

- [提交 Issue](https://github.com/actele/iKuaiAuth/issues)
- [查看主文档](../README.md)
