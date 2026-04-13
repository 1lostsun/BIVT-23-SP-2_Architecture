FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ВАЖНО: Явно указываем GOOS=linux и GOARCH=amd64 (или arm64 для совместимости)
# Для большинства Docker хабов нужен linux/amd64
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o server .

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/server .

RUN chmod +x /app/server

EXPOSE 8080

CMD ["./server"]