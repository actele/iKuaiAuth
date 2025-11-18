# iKuai 认证服务文档索引 📚

欢迎查阅 iKuai 认证服务的完整文档！

## 📖 快速导航

### 新手入门

1. **[README.md](../README.md)** - 项目概述和快速开始
   - 核心特性
   - 快速安装
   - 基础配置
   - iKuai 路由器配置

### 部署和配置

2. **[DEPLOYMENT.md](DEPLOYMENT.md)** - 部署指南
   - Ubuntu/Debian 部署
   - CentOS/RHEL 部署
   - Docker 部署
   - Nginx 反向代理
   - 服务管理
   - 防火墙配置
   - 故障排查

3. **[API_AUTH.md](API_AUTH.md)** - API 认证配置
   - 认证模式对比
   - API 配置方法
   - 请求和响应格式
   - 配置示例
   - 测试验证
   - 示例实现（Python/Node.js/PHP）

### 开发和维护

4. **[ARCHITECTURE.md](ARCHITECTURE.md)** - 系统架构
   - 设计理念
   - 技术栈
   - 模块设计
   - 认证流程
   - Token 计算
   - 前端设计
   - 安全设计

5. **[TESTING.md](TESTING.md)** - 测试指南
   - 快速测试
   - 功能测试
   - API 测试
   - Token 验证
   - 故障排查
   - 性能测试

6. **[CONTRIBUTING.md](../CONTRIBUTING.md)** - 贡献指南
   - 如何贡献
   - 代码规范
   - 开发环境
   - Pull Request 流程
   - 行为准则

## 🎯 按需查阅

### 我想...

#### 快速部署服务

→ [README.md](../README.md) → [DEPLOYMENT.md](DEPLOYMENT.md)

#### 配置外部 API 认证

→ [API_AUTH.md](API_AUTH.md)

#### 了解系统工作原理

→ [ARCHITECTURE.md](ARCHITECTURE.md)

#### 测试和调试

→ [TESTING.md](TESTING.md)

#### 贡献代码

→ [CONTRIBUTING.md](../CONTRIBUTING.md)

## 📋 文档清单

| 文档 | 主要内容 | 适用人群 |
|------|----------|----------|
| **README.md** | 项目介绍、快速开始 | 所有用户 |
| **DEPLOYMENT.md** | 部署和服务管理 | 运维人员 |
| **API_AUTH.md** | API 认证配置 | 集成开发者 |
| **ARCHITECTURE.md** | 技术架构设计 | 开发者 |
| **TESTING.md** | 测试和故障排查 | 开发者、运维人员 |
| **CONTRIBUTING.md** | 贡献指南 | 开源贡献者 |

## 🔍 常见场景

### 场景 1: 首次部署

1. 阅读 [README.md](../README.md) 了解项目
2. 按照 [DEPLOYMENT.md](DEPLOYMENT.md) 部署服务
3. 参考 [TESTING.md](TESTING.md) 验证安装

### 场景 2: 集成企业认证系统

1. 阅读 [API_AUTH.md](API_AUTH.md) 了解 API 认证
2. 按示例实现认证接口
3. 参考 [TESTING.md](TESTING.md) 测试集成

### 场景 3: 二次开发

1. 阅读 [ARCHITECTURE.md](ARCHITECTURE.md) 了解架构
2. 参考 [CONTRIBUTING.md](../CONTRIBUTING.md) 设置开发环境
3. 按照代码规范进行开发

### 场景 4: 故障排查

1. 查看 [TESTING.md](TESTING.md) 的故障排查章节
2. 启用调试模式（见 [README.md](../README.md)）
3. 参考 [DEPLOYMENT.md](DEPLOYMENT.md) 检查服务状态

## 💡 提示

- 📌 **配置文件示例**: 所有文档都包含完整的配置示例
- 🔧 **调试模式**: 遇到问题时，启用 `debug: true` 查看详细日志
- 📝 **日志查看**: 使用 `sudo journalctl -u ikuai-auth -f` 查看实时日志
- 🆘 **获取帮助**: [GitHub Issues](https://github.com/actele/iKuaiAuth/issues)

## 📞 支持

- **Bug 报告**: [提交 Issue](https://github.com/actele/iKuaiAuth/issues)
- **功能建议**: [提交 Issue](https://github.com/actele/iKuaiAuth/issues)
- **贡献代码**: [提交 Pull Request](https://github.com/actele/iKuaiAuth/pulls)

## 📅 文档更新

文档持续更新中，欢迎贡献！

- **最后更新**: 2025-01-05
- **版本**: 1.0.0

---

⭐ 如果觉得有帮助，请给项目一个 Star！
