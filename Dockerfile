# ────────────────────────────────────────────────────────────────────────────
# Multi-stage Dockerfile для device-service
# ────────────────────────────────────────────────────────────────────────────

# 1) Build-stage: компилируем Go-бинарь
FROM golang:1.24-alpine AS builder

# Если в импортах есть зависимости по git, его нужно установить
RUN apk add --no-cache git

WORKDIR /app

# 1.1) Сначала копируем только файлы зависимостей, чтобы кэшировать 'go mod download'
COPY go.mod go.sum ./
RUN go mod download

# 1.2) Копируем весь код и JSON-ключи
COPY . .

# 1.3) Собираем статический бинарь «app»
#      (go build найдёт main.go в корне проекта)
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

# ────────────────────────────────────────────────────────────────────────────
# 2) Runtime-stage: минимальный distroless для запуска
FROM gcr.io/distroless/base-debian11

WORKDIR /app

# 2.1) Копируем собранный бинарь и ключи для Firebase из builder-stage
COPY --from=builder /app/app /app/app
COPY --from=builder /app/service-account.json /app/service-account.json
# Если вы пользуетесь firebase_key.json, скопируйте и его:
# COPY --from=builder /app/firebase_key.json /app/firebase_key.json

# 2.2) Говорим Cloud Run: приложение слушает порт 8080
EXPOSE 8080
ENV PORT=8080

# 2.3) Указываем Firebase SDK, где искать учётные данные
ENV GOOGLE_APPLICATION_CREDENTIALS="/app/service-account.json"

# 2.4) ENTRYPOINT: запускаем Go-бинарь
ENTRYPOINT ["/app/app"]
