# 构建阶段
FROM golang:1.23 AS build

WORKDIR /app

# 复制依赖文件并下载模块
COPY go.mod go.sum ./
RUN go mod download

# 复制项目源代码并构建可执行文件
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 生产阶段 - 使用 distroless 镜像
FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=build /app/main .

# 暴露端口
EXPOSE 8080

# 使用非 root 用户运行
USER nonroot:nonroot

# 设置启动命令
CMD ["/app/main"]
