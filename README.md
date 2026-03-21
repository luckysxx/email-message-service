# Email Message Service 📧

异步邮件发送微服务。通过监听 Kafka 消息队列（Topic: `user.registered`），在用户注册后自动发送欢迎邮件。

## 技术栈
- **Go 1.25+** / **Viper** 配置管理 / **Zap** 结构化日志
- **Kafka-go** 消息消费 / **gomail** SMTP 发送
- **errgroup** 并发与优雅停机 / **Docker** 容器化部署

## 目录结构
```text
├── cmd/main.go              # 主入口：配置加载、依赖注入、生命周期管理
├── config/config.go         # Viper 配置结构体 + godotenv 加载
├── internal/
│   ├── event/               # 事件模型定义 (UserRegisteredEvent)
│   ├── handler/consumer.go  # Kafka 消费者：拉取消息 → 解析 → 调用发件
│   └── service/email.go     # SMTP 发件器：构造邮件 → DialAndSend
├── config.yaml              # 非敏感配置骨架（提交到 Git）
├── .env                     # 敏感凭证（不提交，见 .env.example）
└── docker-compose.yml       # 容器编排
```

## 快速开始

### 1. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env，填入真实的 SMTP 邮箱授权码
```

### 2. 本地运行
```bash
go mod tidy
go run cmd/main.go
```

### 3. Docker 部署
```bash
docker-compose up -d --build

# 查看日志
docker logs -f email-message-service
```

## 配置说明

### config.yaml（非敏感，提交到 Git）
| 字段 | 说明 | 示例 |
|------|------|------|
| `app.env` | 运行环境 | `development` |
| `kafka.brokers` | Kafka 地址 | `global-kafka:9092` |
| `kafka.topic` | 消费的主题 | `user.registered` |
| `smtp.host` | SMTP 服务器 | `smtp.163.com` |
| `smtp.port` | SMTP 端口 | `25` |

### .env（敏感，不提交）
| 变量 | 说明 |
|------|------|
| `APP_ENV` | 运行环境，影响日志颜色 |
| `SMTP_USERNAME` | 邮箱账号 |
| `SMTP_PASSWORD` | 邮箱 SMTP 授权码 |
| `SMTP_FROM` | 发件人地址 |
