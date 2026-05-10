# Сборка
FROM golang:1.26.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o dungeon_app ./cmd/app/main.go

# Запуск 
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/dungeon_app .

RUN adduser -D -u 1000 appuser && chown -R appuser /app
USER appuser

ENTRYPOINT ["./dungeon_app"]