# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.26-alpine AS builder

ARG TARGETARCH
ARG TARGETOS=linux

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 缓存依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译二进制
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /app/server main.go

# ============================================
# Stage 2: Runtime
# ============================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata bash \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从 builder 阶段复制二进制
COPY --from=builder /app/server /app/server

# 复制配置文件
COPY manifest/config /app/manifest/config

# 暴露服务端口
EXPOSE 8000

# 启动服务
ENTRYPOINT ["/app/server"]

