# Email Message Service 📧

这是一个轻量但符合企业级微服务规范的邮件发送服务。它作为底层解耦的基础设施，通过异步监听 Kafka 消息队列（Topic: `user.registered`）来实现邮件群发与通知。

## 🚀 核心技术栈
- **语言**：[Go 1.25+](https://golang.google.cn/)
- **配置中心**：[Viper](https://github.com/spf13/viper)
- **结构化追踪日志**：[Zap Logger (Customized)](https://github.com/uber-go/zap)
- **并发与生命周期控制**： `golang.org/x/sync/errgroup` (优雅启停)
- **消息队列驱动**：[Kafka-go](https://github.com/segmentio/kafka-go)
- **容器化部署**：Multi-stage Docker, Docker Compose

---

## 📂 目录结构规范
```text
.
├── cmd/
│   └── main.go                 # 进程主入口：负责启动环境与统一的依赖倒置(DI)组装
├── config/
│   └── config.go               # 配置结构体映射及自动解析
├── internal/
│   ├── event/                  # Event/Message 模型定义
│   ├── handler/                # Kafka 消费端核心逻辑 (无缝对接底层服务)
│   └── service/                # 具体业务逻辑及 SMTP 发件协议封装
├── pkg/
│   └── logger/                 # 跨项目复用的日志基础设施，完美区分 Dev / Prod
├── Dockerfile                  # 极简容器化编译文件
├── docker-compose.yml          # 集成部署编排 (挂载外部网络)
└── config.example.yaml         # 配置文件范例模板
```

---

## 🛠️ 如何运行

### 1. 准备配置文件
因为 `config.yaml` 中包含真实的邮箱密码和授权码等敏感信息，所以它**已被加入 `.gitignore`，不会被提交到代码仓库**。

请在项目根目录复制一份模板作为你的真实配置：
```bash
cp config.example.yaml config.yaml
```
然后打开 `config.yaml` 填入你的**真实邮箱与授权码**。

### 2. 本地开发调试 (Local Run)
确保你的 Kafka 在 `localhost:9092` 可用，并将 `config.yaml` 中的 brokers 指向本地。
```bash
# 自动整理 Go 依赖
go mod tidy

# 编译并运行
go run cmd/main.go
```

### 3. 使用 Docker 部署 (Production/Stack)
该服务已配置好接入外部的 `go-net` 基础设施网络（包含 Kafka 和数据库集群等）。
确保 `config.yaml` 的 Kafka 地址为 `global-kafka:9092`，然后执行：
```bash
docker-compose up -d --build
```

通过以下命令查看精美的控制台日志流：
```bash
docker logs -f email-message-service
```
