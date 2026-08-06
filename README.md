# iKuai 认证服务 🚀

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/actele/iKuaiAuth/pulls)

> 基于 Go 语言的现代化 iKuai 路由器认证系统，支持本地用户和外部 API 双认证模式

## ✨ 核心特性

- 🎨 **Ant Design 白色主题** - 简洁现代的用户界面，支持暗黑模式切换
- � **双认证模式** - 支持本地用户配置和外部 API 认证
- �🚀 **单文件部署** - 所有静态资源嵌入二进制文件，零依赖部署
-  **响应式设计** - 完美适配移动端和桌面端
- ⚡ **高性能** - Go 语言编写，内存占用低，启动迅速
- 🛠️ **灵活配置** - YAML 配置文件，支持调试模式
- 🎯 **美化 URL** - 无 `.html` 后缀的优雅路由
- 🏠 **导航页面** - 集成搜索引擎和 AI 工具链接

## 📸 界面预览

- **登录页面**: Ant Design 白色主题，简洁大方
- **成功页面**: 自动倒计时（生产模式）/ 调试信息（调试模式）
- **导航页面**: 时间显示 + 搜索引擎 + AI 工具链接

## 🚀 快速开始

### 前置要求

- Go 1.23+ （如果需要从源码编译）
- iKuai 路由器并获取 AppKey

### 安装部署

#### 方式 1: 使用预编译二进制（推荐）

```bash
# 1. 下载最新 release
wget https://github.com/actele/iKuaiAuth/releases/latest/download/ikuai-auth-linux-amd64.tar.gz

# 2. 解压
tar -xzf ikuai-auth-linux-amd64.tar.gz
cd ikuai-auth-linux-amd64

# 3. 配置服务
cp config.yaml.example config.yaml
nano config.yaml  # 修改配置

# 4. 运行安装脚本（Ubuntu/Debian）
chmod +x install-ubuntu.sh
sudo ./install-ubuntu.sh
```

#### 方式 2: 从源码编译

```bash
# 1. 克隆仓库
git clone https://github.com/actele/iKuaiAuth.git
cd iKuaiAuth

# 2. 安装依赖
go mod tidy

# 3. 配置
cp config.yaml.example config.yaml
nano config.yaml

# 4. 编译运行
go build -o ikuai-auth
./ikuai-auth
```

### 基础配置

编辑 `config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: "8088"
  debug: false  # 生产环境设为 false

ikuai:
  portal_url: "https://portal.ikuai8-wifi.com/Action/webauth-up"
  app_key: "your_app_key_here"  # 从爱快云平台获取

auth:
  method: "simple"  # 或 "api"
  simple_users:
    user01: "password123"
```

### iKuai 路由器配置

1. 登录爱快路由器管理界面
2. 进入 **网络设置** → **认证配置**
3. 选择 **自定义认证**
4. 填写认证 URL: `http://your-server-ip:8088/`
5. 保存配置

## 📖 文档

- [部署指南](docs/DEPLOYMENT.md) - Ubuntu 系统部署和服务管理
- [API 认证配置](docs/API_AUTH.md) - 外部 API 认证对接
- [系统架构](docs/ARCHITECTURE.md) - 技术架构和设计理念
- [测试指南](docs/TESTING.md) - 功能测试和故障排查

## � 认证模式

### 本地用户认证（Simple）

```yaml
auth:
  method: "simple"
  simple_users:
    admin: "admin123"
    user01: "password456"
```

### API 认证（推荐）

```yaml
auth:
  method: "api"
  api:
    url: "https://your-api.com/auth/verify"
    method: "POST"
    timeout: 5
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer your-token"
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

详见 [API 认证配置文档](docs/API_AUTH.md)

## 📁 项目结构

```
iKuaiAuth/
├── main.go              # 应用入口
├── router.go            # 路由配置
├── config/              # 配置模块
│   └── config.go       # 配置结构定义
├── handlers/            # 请求处理
│   └── auth.go         # 认证处理器
├── utils/               # 工具函数
│   ├── token.go        # MD5 token 计算
│   ├── api_auth.go     # API 认证
│   └── logger.go       # 日志工具
├── web/                 # 前端资源（嵌入）
│   ├── index.html      # 登录页
│   ├── success.html    # 成功页
│   ├── nav.html        # 导航页
│   └── auth.js         # 认证逻辑
├── docs/                # 项目文档
│   ├── DEPLOYMENT.md   # 部署指南
│   ├── API_AUTH.md     # API 认证配置
│   ├── NETWORK_AUTH.md # 网络设备认证
│   ├── ARCHITECTURE.md # 系统架构
│   ├── TESTING.md      # 测试指南
│   └── CONTRIBUTING.md # 贡献指南
└── scripts/             # 构建和部署脚本
    ├── build-ubuntu.sh # Ubuntu 编译脚本
    ├── build-embedded.sh # 嵌入式编译脚本
    └── start.sh        # 启动脚本
```

## �️ 调试模式

开发时启用调试模式：

```yaml
server:
  debug: true
```

**调试模式特性：**
- ✅ 显示详细的认证参数
- ✅ 显示生成的放行 URL
- ✅ 手动点击"开始上网"按钮
- ✅ 跳转到 iKuai 页面查看响应

**生产模式特性：**
- ✅ 静默认证，无调试信息
- ✅ 自动 3 秒倒计时
- ✅ iframe 静默放行请求
- ✅ 自动跳转到导航页

## 🔒 安全建议

1. ✅ `config.yaml` 已自动排除（`.gitignore`）
2. ✅ 使用强密码或集成外部认证系统
3. ✅ 生产环境使用 Nginx 反向代理
4. ✅ 配置 HTTPS（使用 Let's Encrypt）
5. ✅ 定期更新依赖和服务

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

查看 [贡献指南](docs/CONTRIBUTING.md) 了解更多

## 📄 许可证

本项目基于 [MIT License](LICENSE) 开源

## 🙏 鸣谢

- [iKuai 路由器](https://www.ikuai8.com/) - 官方文档和技术支持
- [Ant Design](https://ant.design/) - UI 设计系统
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP 路由框架
- [go-yaml](https://github.com/go-yaml/yaml) - YAML 解析库

## 📞 支持

- 🐛 [提交 Bug](https://github.com/actele/iKuaiAuth/issues)
- 💡 [功能建议](https://github.com/actele/iKuaiAuth/issues)
- 📖 [查看文档](docs/)

---

⭐ 如果这个项目对您有帮助，请给它一个 Star！
