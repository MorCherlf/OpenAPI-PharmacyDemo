FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o main .

# 使用 distroless 镜像
FROM gcr.io/distroless/base-debian10

WORKDIR /root/
COPY --from=build /app/main .

EXPOSE 8080
CMD ["/root/main"]
