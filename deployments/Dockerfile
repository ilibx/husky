# 构建阶段
#
# internal/admin/admin.go 用 //go:embed static/* 把前端打进二进制，
# 所以构建镜像时必须先产出前端文件，再执行 go build。

# ---- 前端构建 ----
FROM node:20-alpine AS web

WORKDIR /web

# 先只拷依赖清单，让 npm ci 这一层可缓存
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund

COPY web/ ./
RUN npx vite build --mode embed

# ---- Go 构建 ----
# 版本需 >= go.mod 声明的 go 1.24.0
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装必要的依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum* ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# go:embed 会跳过 "_" 开头的文件，而 Vite 产物里有
# _plugin-vue_export-helper-*.js 这类文件名。统一走 prepare-admin-embed.sh
# 做重命名 + 引用替换，与 make web-build 共用同一份逻辑。
COPY --from=web /web/dist ./web/dist
RUN sh scripts/prepare-admin-embed.sh web/dist internal/admin/static

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./main"]
