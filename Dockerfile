# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o monolith ./cmd/server

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/monolith .
COPY .env .

EXPOSE 3000

CMD ["./monolith"]