# Runtime stage
FROM golang:1.24-alpine AS builder

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Скопировать файлы модулей, чтобы rootpath их увидел
COPY --from=builder /app/go.mod .
COPY --from=builder /app/go.sum .

# Копируем бинарник
COPY --from=builder /app/tasker .

EXPOSE 3000

CMD ["./tasker"]
