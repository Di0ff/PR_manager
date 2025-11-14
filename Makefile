# ==============================
# Makefile — Docker-first approach
# ==============================

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# ----------- Variables ----------
APP_NAME        ?= pr-service
DOCKER_COMPOSE  := docker-compose
DOCKER_COMPOSE_E2E := docker-compose -f docker-compose.e2e.yaml

.DEFAULT_GOAL := help

.PHONY: help up down restart logs shell db-shell build clean \
        test test-e2e fmt lint mock deps \
        health ps

# ---------- HELP ----------
help: ## Показать список команд
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ---------- Development ----------
up: ## Запустить все сервисы
	$(DOCKER_COMPOSE) up --build

up-d: ## Запустить в фоновом режиме
	$(DOCKER_COMPOSE) up --build -d

down: ## Остановить все сервисы
	$(DOCKER_COMPOSE) down

down-v: ## Остановить + удалить volumes (очистка БД)
	$(DOCKER_COMPOSE) down -v

restart: down up ## Перезапустить все сервисы

logs: ## Показать логи всех сервисов
	$(DOCKER_COMPOSE) logs -f

logs-app: ## Показать только логи приложения
	$(DOCKER_COMPOSE) logs -f app

logs-db: ## Показать только логи БД
	$(DOCKER_COMPOSE) logs -f db

ps: ## Показать запущенные контейнеры
	$(DOCKER_COMPOSE) ps

# ---------- Shell Access ----------
shell: ## Войти в контейнер приложения
	$(DOCKER_COMPOSE) exec app sh

db-shell: ## Войти в PostgreSQL
	$(DOCKER_COMPOSE) exec db psql -U postgres -d name_db

# ---------- Build ----------
build: ## Пересобрать образы
	$(DOCKER_COMPOSE) build

# ---------- Format & Lint ----------
fmt: ## Форматирование кода (в Docker)
	docker run --rm -v $(PWD):/app -w /app golang:1.24 sh -c "go install mvdan.cc/gofumpt@latest && go install golang.org/x/tools/cmd/goimports@latest && gofumpt -w . && goimports -w ."

lint: ## Запуск линтера (в Docker)
	docker run --rm -v $(PWD):/app -w /app golang:1.24 sh -c "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && golangci-lint run"

lint-fix: ## Линтер + автофикс (в Docker)
	docker run --rm -v $(PWD):/app -w /app golang:1.24 sh -c "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && golangci-lint run --fix"

# ---------- Tests ----------
test: ## Запуск всех тестов в Docker
	docker run --rm -v $(PWD):/app -w /app golang:1.24 go test -race -coverprofile=coverage.out ./...

test-unit: ## Unit тесты в Docker
	docker run --rm -v $(PWD):/app -w /app golang:1.24 go test -race ./internal/...

test-e2e: ## End-to-End тесты в Docker
	$(DOCKER_COMPOSE_E2E) up --build --exit-code-from tests
	$(DOCKER_COMPOSE_E2E) down -v

cover: ## Покрытие в консоли
	docker run --rm -v $(PWD):/app -w /app golang:1.24 go tool cover -func=coverage.out

# ---------- Mocks ----------
mock: ## Генерация моков в Docker
	docker run --rm -v $(PWD):/app -w /app golang:1.24 sh -c "go install github.com/vektra/mockery/v2@v2.46.0 && mockery --all --output ./mocks --case underscore"

# ---------- Dependencies ----------
deps: ## Обновить зависимости
	docker run --rm -v $(PWD):/app -w /app golang:1.24 sh -c "go mod tidy && go mod download"

# ---------- Health & Debug ----------
health: ## Проверить health endpoint
	@curl -s http://localhost:8080/health | jq . || echo "Сервис недоступен"

# ---------- Cleaning ----------
clean: ## Остановить все + удалить volumes и артефакты
	$(DOCKER_COMPOSE) down -v
	rm -rf bin coverage.out mocks

clean-all: clean ## Полная очистка (включая Docker образы)
	docker rmi $(APP_NAME):latest pr_manager-app || true
