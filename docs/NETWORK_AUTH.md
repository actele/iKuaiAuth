# 网络认证 API 集成指南 🔐

本文档说明如何将 iKuai 认证系统与网络认证 API 集成，实现基于设备（MAC + IP）的认证。

> **注意**：网络认证不是独立的认证模式，而是 `api` 认证模式的一种应用场景。通过配置 API 认证的请求格式，可以实现网络设备认证功能。

## 目录

- [概述](#概述)
- [认证模式说明](#认证模式说明)
- [网络认证 API 要求](#网络认证-api-要求)
- [配置方法](#配置方法)
- [认证流程](#认证流程)
- [配置示例](#配置示例)
- [测试验证](#测试验证)
- [故障排查](#故障排查)

## 概述

### 什么是网络认证

网络认证是一种基于设备 MAC 地址和 IP 地址的认证方式，适用于需要设备管理和 IP 分配控制的企业网络环境。它通过 API 认证模式实现，使用特定的请求格式与设备管理系统对接。

### 应用场景

- ✅ 企业内网设备管理
- ✅ 固定 IP 分配和审批
- ✅ 设备自动入库
- ✅ MAC 地址白名单管理
- ✅ IP 地址绑定管理

### 认证逻辑

系统支持三种认证场景：

#### 场景一：已在库 + 已分配 IP（自动通过）

**条件**：
- 设备 MAC 地址已在设备库中
- 已为该设备分配固定 IP（DHCP 静态绑定）
- 分配状态为"已审批通过"

**处理流程**：
1. 验证 MAC 地址格式
2. 查询设备是否在库
3. 查询是否有固定 IP 分配记录
4. 检查分配状态是否为"已审批"
5. 验证 IP 地址是否与分配的 IP 一致

**认证结果**：✅ **自动通过认证**

#### 场景二：已在库 + 未分配 IP（自动绑定）

**条件**：
- 设备 MAC 地址已在设备库中
- 未分配固定 IP（或分配记录未审批）

**处理流程**：
1. 验证 MAC 地址格式
2. 查询设备是否在库
3. **自动创建 IP 绑定记录**（MAC 地址 + IP 地址）
4. 设置状态为"待审批"

**认证结果**：❌ **拒绝认证**，提示"IP 绑定已创建，等待管理员审批"

**后续流程**：
- 管理员在后台审批该绑定记录
- 审批通过后，设备下次认证时走"场景一"逻辑，自动通过

#### 场景三：不在库（自动入库 + 绑定）

**条件**：
- 设备 MAC 地址未在设备库中
- 需要用户提供姓名、手机号

**处理流程**：
1. 验证 MAC 地址格式
2. 查询设备不在库
3. 验证必填参数（姓名、手机号）
4. **自动创建设备记录**（MAC 地址、姓名、手机号）
5. **自动创建 IP 绑定记录**（MAC 地址 + IP 地址）
6. 设置状态为"待审批"

**认证结果**：❌ **拒绝认证**，提示"设备已入库，IP 绑定已创建，等待管理员审批"

**后续流程**：
- 管理员在后台审批该设备和绑定记录
- 审批通过后，设备下次认证时走"场景一"逻辑，自动通过

## 认证模式说明

iKuai 认证系统支持两种认证模式：

| 认证模式 | 说明 | 适用场景 |
|---------|------|---------|
| **simple** | 本地用户认证，基于配置文件中的用户名密码 | 小型部署、测试环境 |
| **api** | 外部 API 认证，调用外部认证系统 | 企业集成、统一认证、设备管理 |

**网络认证** 是 `api` 模式的一种应用场景，通过配置特定的请求格式实现设备认证：

- **标准 API 认证**：传递 username/password，用于用户登录
- **网络设备认证**：传递 device_number/mac_address/ip_address，用于设备管理

两者都使用 `method: "api"` 配置，区别仅在于 `body_template` 的格式不同。

## 网络认证 API 要求

### API 接口规范

**接口地址**：`POST /api/v1/network-auth/authenticate`

### 请求格式

```json
{
  "mac_address": "AA:BB:CC:DD:EE:FF",
  "ip_address": "192.168.1.100",
  "system_id": 1,
  "name": "张三",
  "phone": "13900139000"
}
```

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| mac_address | string | 是 | 设备 MAC 地址（格式：AA:BB:CC:DD:EE:FF，大写） |
| ip_address | string | 是 | 设备当前 IP 地址 |
| system_id | integer | 是 | 路由系统 ID |
| name | string | 否 | 用户姓名（MAC 不在库时必填） |
| phone | string | 否 | 手机号（MAC 不在库时必填） |

### 响应格式

```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "allowed": true,
    "reason": "认证通过",
    "device_id": "device-uuid-123",
    "device_name": "小米13 Pro",
    "assigned_ip": "192.168.1.100",
    "assigned_to": "张三",
    "user_id": 123,
    "assignment_id": "assignment-uuid-456"
  },
  "time": 1700123456
}
```

### 响应字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| allowed | boolean | 是否允许通过认证（true=通过，false=拒绝） |
| reason | string | 认证结果描述 |
| device_id | string | 设备唯一标识 |
| device_name | string | 设备名称 |
| assigned_ip | string | 分配的固定 IP 地址 |
| assigned_to | string | 分配给的用户姓名 |
| user_id | integer | 用户 ID |
| assignment_id | string | IP 分配记录 ID |

## 配置方法

### 1. 修改配置文件

编辑 `config.yaml`，使用 `api` 模式并配置网络认证格式：

```yaml
auth:
  method: "api"  # 使用 API 认证模式
  
  api:
    url: "http://192.168.1.100:8000/api/v1/network-auth/authenticate"
    method: "POST"
    timeout: 5
    
    # 自定义请求头
    headers:
      Content-Type: "application/json"
      X-API-Version: "v1"
    
    # 网络设备认证请求格式
    body_template: |
      {
        "device_number": "",
        "user_account": "{{username}}",
        "vpn_number": "",
        "mac_address": "{{password}}",
        "ip_address": "auto",
        "system_id": 1
      }
    
    # 响应验证
    response:
      success_field: "code"
      success_value: 0
      message_field: "msg"
```

### 2. 配置参数说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| url | string | ✅ 是 | 网络认证 API 的完整 URL |
| method | string | ✅ 是 | HTTP 方法，固定为 "POST" |
| timeout | int | ⚪ 否 | 超时时间（秒），默认 5 |
| headers | map | ⚪ 否 | 自定义 HTTP 请求头 |
| body_template | string | ✅ 是 | 请求体模板，使用 {{username}} 和 {{password}} 变量 |
| response | object | ✅ 是 | 响应验证配置 |

**模板变量说明**：
- `{{username}}`：前端输入的用户账号
- `{{password}}`：前端输入的密码（这里用作 MAC 地址）
- `"auto"`：表示自动获取客户端 IP

### 3. 重启服务

```bash
sudo systemctl restart ikuai-auth
```

## 认证流程

### 完整流程图

```
用户连接 WiFi
   │
   ▼
iKuai 路由器拦截
   │
   ▼
重定向到认证页面
   │
   ▼
用户查看设备信息（MAC、IP）
   │
   ▼
点击认证 → 发送请求
   │
   ▼
iKuai 认证服务
   ├── 格式化 MAC 地址（转大写，添加冒号）
   └── 调用网络认证 API
       │
       ▼
网络认证 API
   ├── 场景一：已在库 + 已分配 IP → 允许
   ├── 场景二：已在库 + 未分配 IP → 自动绑定 → 拒绝
   └── 场景三：不在库 → 自动入库 + 绑定 → 拒绝
       │
       ▼
返回认证结果
   │
   ├─→ 允许：生成 iKuai 放行 URL → 放行
   └─→ 拒绝：显示提示信息 → 等待审批
```

### MAC 地址格式化

系统会自动处理 MAC 地址格式：

```
输入格式                   输出格式
aabbccddeeff      →     AA:BB:CC:DD:EE:FF
aa:bb:cc:dd:ee:ff →     AA:BB:CC:DD:EE:FF
aa-bb-cc-dd-ee-ff →     AA:BB:CC:DD:EE:FF
AA:BB:CC:DD:EE:FF →     AA:BB:CC:DD:EE:FF （不变）
```

## 配置示例

### 基础配置

```yaml
server:
  host: "0.0.0.0"
  port: "8088"
  debug: false

ikuai:
  portal_url: "https://portal.ikuai8-wifi.com/Action/webauth-up"
  app_key: "your_app_key_here"

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

### 带自定义请求头

```yaml
auth:
  method: "api"
  
  api:
    url: "http://api.company.com/network-auth/authenticate"
    method: "POST"
    timeout: 10
    
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer sk-1234567890abcdef"
      X-API-Key: "company-internal-key"
    
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

## 测试验证

### 1. 启用调试模式

```yaml
server:
  debug: true
```

### 2. 查看日志

```bash
sudo journalctl -u ikuai-auth -f
```

日志输出示例：

```
网络认证通过 - MAC: AA:BB:CC:DD:EE:FF, IP: 192.168.1.100, 原因: 认证通过
认证成功 - 用户: , IP: 192.168.1.100, MAC: AA:BB:CC:DD:EE:FF
```

### 3. 测试场景

#### 测试场景一：已审批设备

1. 确保设备已在系统中并审批通过
2. 连接 WiFi，触发认证
3. 应该自动通过认证并放行

#### 测试场景二：新设备自动入库

1. 使用未入库的设备连接 WiFi
2. 认证页面会提示输入姓名和手机号
3. 提交后，设备自动入库
4. 等待管理员审批后，再次认证即可通过

## 故障排查

### 问题 1: API 请求超时

**症状**：

```
网络认证API调用失败 - MAC: AA:BB:CC:DD:EE:FF, IP: 192.168.1.100, 错误: context deadline exceeded
```

**解决方案**：

```bash
# 1. 测试 API 连通性
curl -v http://your-api-server:8080/api/v1/network-auth/authenticate

# 2. 增加超时时间
# 修改 config.yaml
network_auth:
  timeout: 10  # 从 5 秒改为 10 秒

# 3. 检查网络连接
ping your-api-server
```

### 问题 2: MAC 地址格式错误

**症状**：

```
网络认证拒绝 - MAC: invalid, IP: 192.168.1.100, 原因: MAC地址格式错误
```

**解决方案**：

系统会自动格式化 MAC 地址，确保 iKuai 路由器传递的 MAC 地址参数正确。

### 问题 3: 认证失败但应该通过

**症状**：

设备已在库且已审批，但仍然无法通过认证

**排查步骤**：

```bash
# 1. 启用调试模式
nano /opt/ikuai-auth/config.yaml
# 设置 debug: true

# 2. 重启服务
sudo systemctl restart ikuai-auth

# 3. 查看详细日志
sudo journalctl -u ikuai-auth -f

# 4. 检查 API 返回的 allowed 字段
# 确认是 true 还是 false

# 5. 检查 system_id 是否匹配
# 确认配置文件中的 system_id 与 API 要求一致
```

### 问题 4: 设备不在库提示缺少信息

**症状**：

```
设备不在库中，请提供姓名和手机号进行注册
```

**解决方案**：

这是正常提示。前端页面需要支持用户输入姓名和手机号：

1. 修改 `web/index.html` 添加姓名和手机号输入框
2. 或者提示用户联系管理员手动添加设备

## 前端集成

### 修改登录表单

在 `web/index.html` 中添加姓名和手机号字段（网络认证模式可选）：

```html
<div class="form-group" id="name-group" style="display: none;">
    <label>姓名</label>
    <input type="text" id="name" placeholder="请输入姓名">
</div>

<div class="form-group" id="phone-group" style="display: none;">
    <label>手机号</label>
    <input type="tel" id="phone" placeholder="请输入手机号">
</div>
```

### 认证请求示例

```javascript
const authData = {
    username: '',  // 网络认证模式可为空
    password: '',  // 网络认证模式可为空
    user_ip: userIP,
    mac: macAddress,
    name: document.getElementById('name')?.value,
    phone: document.getElementById('phone')?.value
};

fetch('/api/auth', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(authData)
})
.then(response => response.json())
.then(data => {
    if (data.success) {
        // 认证成功
        window.location.href = '/success';
    } else {
        // 显示错误信息
        alert(data.message);
    }
});
```

## 安全建议

1. **内网部署**
   - 网络认证 API 应部署在内网
   - 通过防火墙限制访问来源

2. **API 鉴权**
   - 使用 API Key 或 Token 认证
   - 配置在 `headers` 中

3. **HTTPS 传输**
   - 生产环境使用 HTTPS
   - 保护数据传输安全

4. **日志审计**
   - 记录所有认证请求
   - 定期审查异常访问

## 相关文档

- [API 认证配置](API_AUTH.md) - 其他 API 认证方式
- [部署指南](DEPLOYMENT.md) - 服务部署和管理
- [测试指南](TESTING.md) - 功能测试和故障排查

## 支持

遇到问题？

- [提交 Issue](https://github.com/actele/iKuaiAuth/issues)
- [查看主文档](../README.md)

---

**更新时间**: 2025-11-18  
**版本**: v1.0.0
