# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tasker ./cmd/server

# Runtime stage
FROM alpine:latest

# Устанавливаем tzdata для работы с часовыми поясами
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/tasker .

EXPOSE 3000

# Запускаем приложение
CMD ["./tasker"]