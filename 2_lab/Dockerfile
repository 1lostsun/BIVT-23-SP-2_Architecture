FROM golang:latest AS builder

# ВАЖНО: устанавливаем правильный модуль
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем из main.go (указываем правильный путь к модулю)
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM alpine:3.19

WORKDIR /app

# Копируем только бинарник
COPY --from=builder /app/server .

# Добавляем ca-certificates для HTTPS (если нужно)
RUN apk --no-cache add ca-certificates

EXPOSE 8080

CMD ["./server"]