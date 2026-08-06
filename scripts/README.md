# 构建和部署脚本

本目录包含用于构建和部署 iKuai 认证服务的脚本。

## 📄 脚本列表

### 构建脚本

- **`build-ubuntu.sh`** - Ubuntu/Debian 交叉编译脚本
  - 编译 Linux x86_64 版本
  - 创建部署压缩包
  - 包含必要的配置文件

- **`build-embedded.sh`** - 嵌入式资源编译脚本
  - 将 web 静态资源嵌入到二进制文件
  - 生成单文件可执行程序

### 运行脚本

- **`start.sh`** - 服务启动脚本
  - 检查端口占用
  - 后台运行服务
  - 创建 PID 文件

## 🚀 使用方法

### 构建 Ubuntu 版本

```bash
cd scripts
chmod +x build-ubuntu.sh
./build-ubuntu.sh
```

输出文件位于 `dist/` 目录：
- `ikuai-auth` - 可执行文件
- `ikuai-auth-linux-amd64.tar.gz` - 部署压缩包

### 启动服务

```bash
cd scripts
chmod +x start.sh
./start.sh
```

### 停止服务

```bash
pkill -f ikuai-auth
# 或者
kill $(cat ikuai-auth.pid)
```

## 📝 注意事项

1. **构建脚本**：需要在项目根目录的 `scripts/` 下执行
2. **权限**：首次使用需要添加执行权限 `chmod +x *.sh`
3. **环境**：确保已安装 Go 1.23+ 环境
4. **配置**：部署前记得修改 `config.yaml` 中的 `app_key`

## 🔗 相关文档

- [部署指南](../docs/DEPLOYMENT.md) - 详细的部署文档
- [配置示例](../config.yaml.example) - 配置文件模板
