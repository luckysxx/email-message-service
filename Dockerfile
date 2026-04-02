# syntax=docker/dockerfile:1
# ======== 构建阶段 ========
FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o email-service ./cmd

# ======== 运行阶段 ========
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/email-service .

USER appuser

CMD ["./email-service"]
