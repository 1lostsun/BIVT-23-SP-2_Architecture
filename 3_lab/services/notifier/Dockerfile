FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o notifier .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/notifier .
CMD ["./notifier"]
