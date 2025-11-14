# PR Reviewer Assignment Service

Сервис автоматического назначения ревьюверов для Pull Request'ов с учетом командной структуры и активности участников.

## Описание

Сервис реализует систему управления командами, пользователями и Pull Request'ами с автоматическим назначением ревьюверов. При создании PR система автоматически выбирает до 2 активных ревьюверов из команды автора, исключая самого автора и учитывая текущую нагрузку.

### Основные возможности

- Управление командами и участниками
- Автоматическое назначение до 2 ревьюверов при создании PR
- Интеллектуальный выбор ревьюверов с учетом нагрузки (количество активных ревью)
- Переназначение ревьюверов на других членов команды
- Управление активностью пользователей (админ-функция)
- Отслеживание PR'ов назначенных пользователю

## Технологический стек

- **Go 1.24** - язык программирования
- **Gin** - веб-фреймворк
- **PostgreSQL 16** - база данных
- **GORM** - ORM
- **Docker & Docker Compose** - контейнеризация
- **golang-migrate** - миграции БД
- **mockery** - генерация моков для тестов
- **golangci-lint** - статический анализ кода

## Быстрый старт

### Предварительные требования

- Docker и Docker Compose
- Go 1.24+ (для локальной разработки)
- Make (опционально)

### Запуск в Docker

1. Клонировать репозиторий и перейти в директорию:
```bash
  cd PR_manager
```

2. Создать .env файл на основе примера:
```bash
  cp .env.example .env
```

3. Запустить сервис:
```bash
  docker-compose up --build
```

API будет доступен по адресу: `http://localhost:8080`

## API

### Teams

#### POST /team/add
Создать команду с участниками (создаёт или обновляет пользователей).

```bash
  curl -X POST http://localhost:8080/team/add \
    -H "Content-Type: application/json" \
    -d '{
      "team_name": "backend",
      "members": [
        {"user_id": "u1", "username": "Alice", "is_active": true},
        {"user_id": "u2", "username": "Bob", "is_active": true},
        {"user_id": "u3", "username": "Charlie", "is_active": true}
      ]
    }'
```

#### GET /team/get
Получить команду с участниками.

```bash
  curl "http://localhost:8080/team/get?team_name=backend"
```

### Users

#### POST /users/setIsActive
Установить флаг активности пользователя (требуется admin токен).

```bash
  curl -X POST http://localhost:8080/users/setIsActive \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer secret_token" \
    -d '{
      "user_id": "u2",
      "is_active": false
    }'
```

#### GET /users/getReview
Получить PR'ы, где пользователь назначен ревьювером.

```bash
  curl "http://localhost:8080/users/getReview?user_id=u2"
```

### Pull Requests

#### POST /pullRequest/create
Создать PR с автоматическим назначением ревьюверов.

```bash
  curl -X POST http://localhost:8080/pullRequest/create \
    -H "Content-Type: application/json" \
    -d '{
      "pull_request_id": "pr-1001",
      "pull_request_name": "Add search feature",
      "author_id": "u1"
    }'
```

#### POST /pullRequest/merge
Пометить PR как MERGED (идемпотентная операция).

```bash
  curl -X POST http://localhost:8080/pullRequest/merge \
    -H "Content-Type: application/json" \
    -d '{
      "pull_request_id": "pr-1001"
    }'
```

#### POST /pullRequest/reassign
Переназначить ревьювера на другого члена команды.

```bash
  curl -X POST http://localhost:8080/pullRequest/reassign \
    -H "Content-Type: application/json" \
    -d '{
      "pull_request_id": "pr-1001",
      "old_user_id": "u2"
    }'
```

### Health Check

#### GET /health
Проверка работоспособности сервиса.

```bash
  curl http://localhost:8080/health
```

## Тестирование

### Unit тесты
```bash
  make test-unit
```

### E2E тесты
```bash
  make test-e2e
```

### Линтинг
```bash
  make lint
```

## Архитектурные решения

### Структура проекта

```
PR_manager/
├── cmd/                  # Точка входа
│   └── main.go
├── db/scripts/           # SQL миграции
├── internal/             
│   ├── api/              # HTTP слой
│   │   ├── handlers/     
│   │   └── routers/      
│   ├── config/           # Конфигурация
│   ├── custom/           # Кастомные ошибки
│   ├── logger/           # Логирование
│   ├── service/          # Бизнес-логика
│   └── storage/          # Слой данных
│       ├── models/       
│       └── repository/   
└── docker-compose.yml   # Docker конфигурация
```