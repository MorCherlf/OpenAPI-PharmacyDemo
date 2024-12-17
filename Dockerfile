FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=build /app/main .

EXPOSE 8080
CMD ["./main"]
