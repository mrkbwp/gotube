FROM golang:1.22.10-alpine

WORKDIR /app

# Установка необходимых пакетов
RUN apk add --no-cache \
    ffmpeg \
    postgresql-client

# Создаем директорию для временных файлов
RUN mkdir -p /tmp/video-conversion && \
    chmod 777 /tmp/video-conversion

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd/api

EXPOSE 8111

CMD ["./main"]
