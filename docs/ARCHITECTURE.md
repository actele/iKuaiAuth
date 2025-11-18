# 系统架构与设计 🏗️

本文档详细介绍 iKuai 认证系统的技术架构、设计理念和实现细节。

## 目录

- [设计理念](#设计理念)
- [系统架构](#系统架构)
- [技术栈](#技术栈)
- [模块设计](#模块设计)
- [认证流程](#认证流程)
- [Token 计算](#token-计算)
- [前端设计](#前端设计)
- [安全设计](#安全设计)

## 设计理念

### 核心目标

1. **简单部署** - 单一可执行文件，嵌入所有静态资源
2. **灵活认证** - 支持本地和外部 API 双认证模式
3. **现代界面** - Ant Design 白色主题，支持暗黑模式
4. **高性能** - 低内存占用，快速响应
5. **易于维护** - 清晰的代码结构，完整的文档

### 设计原则

- ✅ **零依赖部署** - 所有资源嵌入二进制文件
- ✅ **配置驱动** - 通过 YAML 配置文件控制行为
- ✅ **调试友好** - 调试模式提供详细信息
- ✅ **安全第一** - 敏感配置不入库，支持 HTTPS
- ✅ **开源友好** - MIT 许可证，欢迎贡献

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                        用户设备                          │
│                     (手机/电脑)                          │
└────────────────┬────────────────────────────────────────┘
                 │ HTTP Request
                 ▼
┌─────────────────────────────────────────────────────────┐
│                    iKuai 路由器                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │  重定向到认证页面                                 │  │
│  │  http://auth-server:8088/?user_ip=...&mac=...    │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│              iKuai 认证服务 (Go)                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Gorilla Mux Router                              │  │
│  │    ├── / (index.html)                            │  │
│  │    ├── /login (success.html)                     │  │
│  │    ├── /success (success.html)                   │  │
│  │    ├── /nav (nav.html)                           │  │
│  │    └── /api/auth                                 │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Authentication Handler                           │  │
│  │    ├── Simple Mode (本地用户)                     │  │
│  │    └── API Mode (外部 API)                       │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Token Generator                                  │  │
│  │    └── MD5 计算 (iKuai 协议)                     │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Embedded Web Files                              │  │
│  │    ├── index.html (登录页)                       │  │
│  │    ├── success.html (成功页)                     │  │
│  │    └── nav.html (导航页)                         │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────┬────────────────────────────────────────┘
                 │ (API Mode)
                 ▼
┌─────────────────────────────────────────────────────────┐
│              外部认证 API (可选)                         │
│       (您的用户系统/LDAP/数据库)                         │
└─────────────────────────────────────────────────────────┘
```

### 请求流程

```
1. 用户连接 WiFi
   │
   ▼
2. iKuai 路由器拦截
   │
   ▼
3. 重定向到认证页面
   http://auth:8088/?user_ip=...&mac=...
   │
   ▼
4. 展示登录表单
   │
   ▼
5. 用户输入用户名密码
   │
   ▼
6. POST /api/auth
   │
   ├─→ Simple Mode: 检查配置文件
   │
   └─→ API Mode: 调用外部 API
   │
   ▼
7. 验证成功，生成 Token
   │
   ▼
8. 构建放行 URL
   https://portal.ikuai8.com/Action/webauth-up?...&token=xxx
   │
   ├─→ Debug Mode: 显示 URL，手动点击
   │
   └─→ Production: iframe 自动放行
   │
   ▼
9. 重定向到导航页 (/nav)
```

## 技术栈

### 后端

| 技术 | 版本 | 用途 |
|------|------|------|
| **Go** | 1.23+ | 主要编程语言 |
| **Gorilla Mux** | latest | HTTP 路由框架 |
| **go-yaml** | v3 | YAML 配置解析 |
| **embed** | std | 静态文件嵌入 |
| **crypto/md5** | std | Token 计算 |
| **net/http** | std | HTTP 服务和客户端 |

### 前端

| 技术 | 版本 | 用途 |
|------|------|------|
| **Ant Design** | 5.x | UI 设计系统 |
| **原生 JavaScript** | ES6+ | 前端交互逻辑 |
| **CSS3** | - | 样式和动画 |
| **FontAwesome** | 6.x | 图标库 |

### 部署

| 技术 | 用途 |
|------|------|
| **systemd** | Linux 服务管理 |
| **Nginx** | 反向代理（可选） |
| **Let's Encrypt** | SSL 证书（可选） |
| **Docker** | 容器化部署（可选） |

## 模块设计

### 目录结构

```
iKuaiAuth/
├── main.go                  # 应用入口，加载配置，启动服务
├── router.go                # 路由配置，URL 映射
├── config/
│   └── config.go           # 配置结构定义
├── handlers/
│   └── auth.go             # 认证逻辑处理
├── utils/
│   ├── token.go            # Token 生成（MD5）
│   └── api_auth.go         # 外部 API 认证
├── web/                     # 前端资源（嵌入）
│   ├── index.html          # 登录页
│   ├── success.html        # 成功页
│   └── nav.html            # 导航页
├── docs/                    # 文档
├── scripts/                 # 部署脚本
└── config.yaml.example      # 配置模板
```

### 核心模块

#### 1. main.go - 应用入口

```go
功能：
- 加载 config.yaml 配置
- 嵌入 web/ 目录静态文件
- 初始化路由
- 启动 HTTP 服务器

关键代码：
//go:embed web/*
var webFS embed.FS
```

#### 2. router.go - 路由配置

```go
功能：
- 定义 URL 路由规则
- 处理静态文件服务
- 美化 URL（无 .html 后缀）

路由映射：
/           → web/index.html
/login      → web/success.html
/success    → web/success.html
/nav        → web/nav.html
/api/auth   → handlers.HandleAuth
```

#### 3. config/config.go - 配置管理

```go
结构定义：
type Config struct {
    Server ServerConfig
    IKuai  IKuaiConfig
    Auth   AuthConfig
}

配置加载：
- 读取 config.yaml
- 解析 YAML 到结构体
- 验证必填字段
```

#### 4. handlers/auth.go - 认证处理

```go
功能：
- 解析认证请求
- 双模式认证（simple/api）
- 生成放行 URL
- 返回 JSON 响应

认证流程：
1. 提取用户名密码
2. 根据 method 选择认证方式
3. 验证成功生成 Token
4. 构建 iKuai 放行 URL
```

#### 5. utils/token.go - Token 生成

```go
功能：
- 实现 iKuai Token 计算算法
- MD5 哈希生成

参数顺序（重要）：
custom_name, download, key, mac, release_type,
timeout, timestamp, type, upload, user_id, user_ip
```

#### 6. utils/api_auth.go - API 认证

```go
功能：
- 调用外部认证 API
- 模板变量替换
- 响应解析和验证

支持：
- GET/POST 请求
- 自定义 Headers
- 超时控制
- 灵活的响应验证
```

## 认证流程

### Simple 模式（本地认证）

```
用户输入
   │
   ▼
检查 config.yaml 中的 simple_users
   │
   ├─→ 存在且密码匹配 → 认证成功
   │
   └─→ 不存在或密码错误 → 认证失败
```

### API 模式（外部认证）

```
用户输入
   │
   ▼
构建 HTTP 请求
   ├── URL: config.api.url
   ├── Method: config.api.method
   ├── Headers: config.api.headers
   └── Body: 替换 body_template 中的变量
   │
   ▼
发送到外部 API
   │
   ▼
接收 JSON 响应
   │
   ▼
验证响应
   ├── 检查 HTTP 状态码 = 200
   ├── 解析 JSON
   ├── 提取 success_field
   └── 比较 success_value
   │
   ├─→ 匹配 → 认证成功
   │
   └─→ 不匹配 → 认证失败（提取 message_field）
```

## Token 计算

### iKuai Token 算法

根据爱快官方文档，Token 计算步骤：

```go
1. 收集参数
   params := {
       "type": "20",
       "user_id": username,
       "custom_name": username,
       "user_ip": user_ip,
       "timestamp": unix_timestamp,
       "mac": mac_address,
       "timeout": "0",
       "upload": "0",
       "download": "0",
       "release_type": "1",
       "key": app_key,
   }

2. 按键名排序
   keys: custom_name, download, key, mac, release_type,
         timeout, timestamp, type, upload, user_id, user_ip

3. 构建查询字符串
   query := "custom_name=user&download=0&key=xxx&mac=...&..."

4. MD5 哈希
   token := md5(query)

5. 构建最终 URL（不包含 key）
   url := portal_url + "?" + query_without_key + "&token=" + token
```

### 关键点

- ⚠️ `key` 参数仅用于计算 Token，不出现在最终 URL 中
- ⚠️ 参数必须按字母顺序排序
- ⚠️ 所有参数都参与计算，包括空值
- ⚠️ `release_type=1` 表示客户端发起放行

## 前端设计

### 页面结构

#### index.html - 登录页

```
功能：
- 显示登录表单
- 解析 URL 参数（user_ip, mac）
- 处理双问号问题（iKuai 路由器 bug）
- 发送认证请求

特性：
- Ant Design 白色主题
- 响应式布局
- 参数预填充
- 错误提示
```

#### success.html - 成功页

```
功能：
- 显示认证成功信息
- 生产模式：3 秒倒计时 + 自动放行
- 调试模式：显示参数 + 手动点击

模式切换：
- 通过后端返回的 debug 字段控制
- iframe 隐藏放行（生产）
- 直接跳转（调试）

特性：
- 暗黑模式切换
- 自动倒计时
- 放行状态显示
```

#### nav.html - 导航页

```
功能：
- 显示当前时间和问候语
- 搜索引擎快速入口（Baidu/Bing/Google）
- AI 工具链接（8 个中文 AI 平台）

特性：
- 暗黑模式支持
- 时间实时更新
- 搜索引擎切换
- 响应式卡片布局
```

### URL 美化

通过 `router.go` 实现无 `.html` 后缀：

```
原始路径           美化后路径
/web/index.html  → /
/web/success.html → /login, /success
/web/nav.html    → /nav
```

### 暗黑模式

```css
/* 白色主题（默认） */
body {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

/* 暗黑模式 */
body.dark-mode {
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}
```

JavaScript 切换：

```javascript
function toggleDarkMode() {
    document.body.classList.toggle('dark-mode');
    localStorage.setItem('darkMode', 
        document.body.classList.contains('dark-mode'));
}
```

## 安全设计

### 配置安全

```gitignore
# .gitignore
config.yaml      # 敏感配置不入库
*.log            # 日志不入库
dist/            # 包含配置的部署包不入库
```

提供示例配置：

```yaml
# config.yaml.example
ikuai:
  app_key: "your_app_key_here"  # 占位符，不是真实密钥
```

### 密码处理

- ✅ 密码不打印到日志
- ✅ 仅记录认证结果（成功/失败）
- ✅ API 模式使用 HTTPS 传输
- ✅ 支持自定义加密传输

### 生产模式

```yaml
server:
  debug: false  # 关闭详细日志
```

生产模式特性：
- ❌ 不显示调试面板
- ❌ 不输出详细参数
- ✅ 仅记录关键事件
- ✅ 静默放行（iframe）

### API 认证安全

```yaml
auth:
  api:
    url: "https://..."  # 强制 HTTPS
    headers:
      Authorization: "Bearer token"  # API 鉴权
```

建议：
1. 使用 HTTPS 传输
2. 添加 API Key 或 Bearer Token
3. 实现请求频率限制
4. 记录审计日志

## 性能优化

### 嵌入式资源

```go
//go:embed web/*
var webFS embed.FS
```

优势：
- ✅ 单一可执行文件
- ✅ 无需额外文件部署
- ✅ 内存直接读取，速度快
- ✅ 不可篡改

### 资源占用

```
内存使用: ~10-20 MB
CPU 占用: 空闲时 < 1%
启动时间: < 1 秒
```

### HTTP 性能

```go
// 使用标准库 http.Server
server := &http.Server{
    Addr:         config.Server.Host + ":" + config.Server.Port,
    Handler:      router,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

## 扩展性

### 支持的认证方式

当前实现：
- ✅ Simple（本地用户）
- ✅ API（外部 HTTP API）

可扩展：
- ⏳ LDAP
- ⏳ 数据库（MySQL/PostgreSQL）
- ⏳ OAuth2
- ⏳ RADIUS

### 添加新认证方式

```go
// handlers/auth.go
func validateUser(username, password string, config *config.Config) bool {
    switch config.Auth.Method {
    case "simple":
        return validateSimple(username, password, config)
    case "api":
        return validateAPI(username, password, config)
    case "ldap":  // 新增
        return validateLDAP(username, password, config)
    default:
        return false
    }
}
```

## 调试模式

### 开发者功能

```yaml
server:
  debug: true
```

启用后：
- ✅ 显示完整的放行 URL
- ✅ 显示所有参数解析
- ✅ 手动点击"开始上网"
- ✅ 跳转到 iKuai 页面查看响应
- ✅ 控制台详细日志

### 日志输出

```go
if config.Server.Debug {
    log.Printf("认证参数: user=%s, ip=%s, mac=%s", 
        username, userIP, mac)
    log.Printf("生成的Token: %s", token)
    log.Printf("放行URL: %s", releaseURL)
}
```

## 未来规划

### 功能增强

- [ ] 多语言支持（i18n）
- [ ] 访客认证（无密码）
- [ ] 认证日志统计
- [ ] Web 管理界面
- [ ] 用户在线管理

### 性能优化

- [ ] 连接池（API 认证）
- [ ] 缓存机制
- [ ] 请求限流
- [ ] 负载均衡支持

### 安全增强

- [ ] 双因素认证（2FA）
- [ ] 验证码支持
- [ ] IP 黑白名单
- [ ] 会话管理

## 技术债务

当前已知限制：

1. **响应验证** - 不支持嵌套 JSON 字段
2. **Token 刷新** - 不支持自动续期
3. **会话管理** - 无持久化会话
4. **并发限制** - 无请求频率控制

## 贡献指南

欢迎贡献代码！请遵循：

1. **代码风格** - 遵循 Go 官方规范
2. **测试覆盖** - 新功能需添加测试
3. **文档更新** - 更新相关文档
4. **提交信息** - 清晰描述变更内容

详见 [CONTRIBUTING.md](../CONTRIBUTING.md)

## 参考资料

- [iKuai 官方文档](https://www.ikuai8.com/)
- [Gorilla Mux](https://github.com/gorilla/mux)
- [Ant Design](https://ant.design/)
- [Go Embed](https://pkg.go.dev/embed)

---

**维护者**: actele  
**最后更新**: 2025-01-05
