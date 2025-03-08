APP_NAME = gotube
GO = go
BUILD_DIR = build
CONFIG_PATH = ./configs

# Версия приложения
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Переменные сборки
LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

# Основная цель
.PHONY: all
all: clean test build

# Сборка приложения
.PHONY: build
build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api

# Очистка
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Запуск тестов
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Запуск приложения
.PHONY: run
run:
	@echo "Running application..."
	CONFIG_PATH=$(CONFIG_PATH) $(GO) run ./cmd/api

# Запуск миграций
.PHONY: migrate
migrate:
	@echo "Running migrations..."
	@migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

# Откат миграций
.PHONY: migrate-down
migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down 1

# Запуск линтера
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

# Docker
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(APP_NAME):$(VERSION)
