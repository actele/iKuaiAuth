FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ikuai-auth ./server

# 使用多阶段构建减小镜像大小
FROM alpine:latest

# 安装 ca-certificates
RUN apk --no-cache add ca-certificates

# 创建工作目录
WORKDIR /root/

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/ikuai-auth .

# 复制配置文件和静态文件
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/web ./web

# 创建日志目录
RUN mkdir -p logs

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./ikuai-auth"]