# 贡献指南 🤝

感谢您对 iKuai 认证服务的关注！我们欢迎各种形式的贡献。

## 贡献方式

### 报告问题

如果您发现 Bug 或有功能建议：

1. 查看 [Issues](https://github.com/actele/iKuaiAuth/issues) 是否已存在相同问题
2. 如果没有，请创建新的 Issue
3. 提供详细信息：
   - 问题描述
   - 复现步骤
   - 预期行为
   - 实际行为
   - 环境信息（OS、Go 版本等）
   - 日志或截图

### 提交代码

#### 1. Fork 仓库

点击页面右上角的 "Fork" 按钮。

#### 2. 克隆到本地

```bash
git clone https://github.com/YOUR_USERNAME/iKuaiAuth.git
cd iKuaiAuth
```

#### 3. 创建分支

```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/your-bug-fix
```

分支命名规范：
- `feature/xxx` - 新功能
- `fix/xxx` - Bug 修复
- `docs/xxx` - 文档更新
- `refactor/xxx` - 代码重构
- `test/xxx` - 测试相关

#### 4. 进行修改

遵循代码规范（见下方）。

#### 5. 提交更改

```bash
git add .
git commit -m "feat: add new authentication method"
```

提交信息格式：
- `feat:` - 新功能
- `fix:` - Bug 修复
- `docs:` - 文档更新
- `style:` - 代码格式调整
- `refactor:` - 重构
- `test:` - 测试
- `chore:` - 构建/工具变更

#### 6. 推送分支

```bash
git push origin feature/your-feature-name
```

#### 7. 创建 Pull Request

在 GitHub 上点击 "New Pull Request"。

## 代码规范

### Go 代码

遵循 [Effective Go](https://go.dev/doc/effective_go) 规范：

```go
// 好的示例
func ValidateUser(username, password string) (bool, error) {
    if username == "" {
        return false, errors.New("username cannot be empty")
    }
    // ...
    return true, nil
}

// 不好的示例
func validate_user(u string,p string)(bool,error){
    if u=="" {
        return false,errors.New("error")
    }
    return true,nil
}
```

**关键点：**
- 使用 `gofmt` 格式化代码
- 导出的函数/变量使用大写开头
- 添加适当的注释
- 错误处理要明确
- 避免 panic，返回 error

### 前端代码

**JavaScript:**

```javascript
// 好的示例
function handleSubmit(event) {
    event.preventDefault();
    const username = document.getElementById('username').value;
    // ...
}

// 不好的示例
function f(e){
    var u=document.getElementById('username').value
}
```

**CSS:**

```css
/* 好的示例 */
.auth-container {
    max-width: 400px;
    margin: 0 auto;
    padding: 20px;
}

/* 不好的示例 */
.c{max-width:400px;margin:0 auto}
```

### 配置文件

**YAML:**

```yaml
# 好的示例 - 有注释说明
server:
  host: "0.0.0.0"  # 监听所有网卡
  port: "8088"      # HTTP 端口
  debug: false      # 生产环境关闭

# 不好的示例 - 无注释
server:
  host: "0.0.0.0"
  port: "8088"
  debug: false
```

## 开发环境

### 必需工具

- Go 1.23+
- Git
- 文本编辑器（推荐 VS Code）

### 推荐工具

- [golangci-lint](https://golangci-lint.run/) - 代码检查
- [air](https://github.com/cosmtrek/air) - 热重载

### 设置开发环境

```bash
# 1. 安装依赖
go mod download

# 2. 复制配置文件
cp config.yaml.example config.yaml

# 3. 编辑配置
nano config.yaml

# 4. 运行服务
go run main.go router.go

# 5. 访问测试
open http://localhost:8088
```

### 使用 air 热重载

安装：

```bash
go install github.com/cosmtrek/air@latest
```

创建 `.air.toml`：

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ."
bin = "tmp/main"
include_ext = ["go", "yaml", "html"]
exclude_dir = ["tmp", "vendor", "logs"]
delay = 1000
```

运行：

```bash
air
```

## 测试

### 运行测试

```bash
# 所有测试
go test ./...

# 带覆盖率
go test -cover ./...

# 详细输出
go test -v ./...
```

### 编写测试

```go
// utils/token_test.go
package utils

import "testing"

func TestGenerateToken(t *testing.T) {
    params := map[string]string{
        "user_ip": "192.168.1.100",
        "mac":     "00:11:22:33:44:55",
        // ...
    }
    
    token := GenerateToken(params)
    
    if token == "" {
        t.Error("Token should not be empty")
    }
    
    if len(token) != 32 {
        t.Errorf("Token length should be 32, got %d", len(token))
    }
}
```

## 文档

### 更新文档

修改代码时，请同步更新相关文档：

- `README.md` - 主要文档
- `docs/DEPLOYMENT.md` - 部署相关
- `docs/API_AUTH.md` - API 认证
- `docs/ARCHITECTURE.md` - 架构设计

### 文档规范

- 使用 Markdown 格式
- 添加适当的标题层级
- 提供代码示例
- 包含配置示例
- 添加警告和提示

示例：

```markdown
## 配置 HTTPS

编辑 `config.yaml`：

```yaml
security:
  enable_https: true
  cert_file: "cert.pem"
  key_file: "key.pem"
```

**注意**: 证书文件必须是 PEM 格式。

⚠️ **警告**: 私钥文件权限应设为 600。
```

## Pull Request 检查清单

提交 PR 前，请确认：

- [ ] 代码遵循项目规范
- [ ] 添加了必要的测试
- [ ] 所有测试通过
- [ ] 更新了相关文档
- [ ] 提交信息清晰明确
- [ ] 没有合并冲突
- [ ] PR 描述完整（做了什么、为什么、如何测试）

## PR 模板

创建 PR 时，请包含以下信息：

```markdown
## 变更类型
- [ ] Bug 修复
- [ ] 新功能
- [ ] 文档更新
- [ ] 代码重构
- [ ] 性能优化
- [ ] 其他

## 变更描述
简要描述您的变更内容...

## 相关 Issue
Fixes #123

## 测试
描述如何测试您的变更...

## 截图（如适用）
添加相关截图...

## 检查清单
- [ ] 代码遵循项目规范
- [ ] 添加了测试
- [ ] 测试通过
- [ ] 文档已更新
```

## 代码审查

提交 PR 后：

1. 项目维护者会审查您的代码
2. 可能会提出修改建议
3. 请及时回复和更新
4. 通过审查后会合并到主分支

## 行为准则

### 我们的承诺

为了营造开放和友好的环境，我们承诺：

- 使用友好和包容的语言
- 尊重不同的观点和经验
- 优雅地接受建设性批评
- 关注对社区最有利的事情
- 对其他社区成员表示同理心

### 不可接受的行为

- 使用性化语言或图像
- 挑衅、侮辱或贬损性评论
- 公开或私下骚扰
- 未经许可发布他人私人信息
- 其他不道德或不专业的行为

## 许可证

贡献的代码将在 [MIT License](../LICENSE) 下发布。

## 联系方式

- GitHub Issues: [提交问题](https://github.com/actele/iKuaiAuth/issues)
- Email: actele@example.com（替换为实际邮箱）

---

再次感谢您的贡献！🎉
