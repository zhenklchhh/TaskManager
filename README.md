# TaskManager

Система управления задачами с поддержкой планирования, повторных попыток и мониторинга.

## Архитектура

- **Go** - основной язык разработки
- **PostgreSQL** - хранение задач
- **Redis** - очередь сообщений и кэширование
- **Chi** - HTTP router
- **Cron** - планирование задач
- **Prometheus** - сбор метрик
- **Grafana** - визуализация метрик

## Функциональность

- Создание и управление задачами
- Планирование задач по cron выражению
- Автоматические повторные попытки при ошибках
- Истекание срока действия задач
- REST API
- Health checks
- Метрики производительности

## Быстрый запуск

### 1. Запуск зависимостей

```bash
docker-compose up -d
```

### 2. Миграции базы данных

```bash
make migrate-up
```

### 3. Запуск API сервера

```bash
CONFIG_PATH=config/local.yaml go run cmd/api/main.go
```

### 4. Запуск worker (опционально)

```bash
go run cmd/worker/worker.go
```

## API Endpoints

### Health Checks
- `GET /health` - проверка состояния системы
- `GET /ready` - проверка готовности к работе

### Metrics
- `GET /metrics` - метрики в формате Prometheus

### Tasks
- `POST /api/v1/tasks` - создание задачи
- `GET /api/v1/tasks/{id}` - получение задачи по ID

### Пример создания задачи

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Task",
    "type": "email",
    "payload": "{\"to\":\"test@example.com\",\"subject\":\"Hello\",\"body\":\"Test message\"}",
    "cron_expr": "*/5 * * * *"
  }'
```

## Мониторинг

### Запуск мониторинга

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

### Доступ к сервисам мониторинга

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### Команды

```bash
# Запуск сервера
make run-server

# Запуск worker
make run-worker

# Запуск всех миграций
make migrate-up

# Откат всех миграций
make migrate-down

# Линтинг
make lint
```