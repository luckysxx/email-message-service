# syntax=docker/dockerfile:1
# Build stage
FROM golang:1.25.5-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置 GOPROXY 代理，优先走官方代理，失败时再 direct
ENV GOPROXY=https://proxy.golang.org,direct

# 1. 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 2. 复制全部源代码
COPY . .

# 3. 编译 Go 程序 
# CGO_ENABLED=0 确保编译出静态链接的二进制文件，非常适合在瘦容器中运行
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o bin/email-service ./cmd

# Run stage (运行阶段使用极小的 alpine 镜像)
FROM alpine:latest

WORKDIR /app

# 安装根证书（发邮件进行 TLS 加密验证必不可少）和时区数据
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

# 从编译阶段拷贝生成的可执行文件
COPY --from=builder /app/bin/email-service .

# 拷贝配置文件（生产环境一般通过 volume 挂载或纯环境变量，这里为了方便演示一并拷贝）
COPY --from=builder /app/config.yaml .

# 启动微服务
CMD ["./email-service"]
