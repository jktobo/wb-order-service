# Этап 1: Сборка приложения
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
# Это кэшируется и не будет выполняться каждый раз, если go.mod/go.sum не менялись
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код
COPY . .

# Собираем приложение. Флаги убирают отладочную информацию и уменьшают размер бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/app/main.go

# Этап 2: Создание легковесного образа для запуска
FROM alpine:latest

WORKDIR /app

# Копируем только скомпилированный бинарник из этапа сборки
COPY --from=builder /app/main .

# Копируем HTML-файл
COPY web/static /app/web/static

# Указываем, какую команду запустить при старте контейнера
CMD ["./main"]