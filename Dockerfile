# MiroChain Dockerfile
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata

# Создаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mirochain cmd/node/main.go

# Финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Создаем пользователя для безопасности
RUN adduser -D -s /bin/sh mirochain

# Создаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл
COPY --from=builder /app/mirochain .

# Создаем директорию для данных
RUN mkdir -p /app/data && chown -R mirochain:mirochain /app

# Переключаемся на пользователя mirochain
USER mirochain

# Открываем порты
EXPOSE 8080 8081 8082 8083 8084 8085 8086 8087 8088 8089 8090

# Устанавливаем переменные окружения
ENV GIN_MODE=release
ENV DATA_DIR=/app/data

# Команда по умолчанию
CMD ["./mirochain", "-port=8080", "-mining=false", "-data=/app/data"]
