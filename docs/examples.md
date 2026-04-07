# Примеры использования TaskManager

## Создание задач

### Email задача
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Welcome Email",
    "type": "email",
    "payload": "{\"to\":\"user@example.com\",\"subject\":\"Welcome!\",\"body\":\"Hello!\"}",
    "cron_expr": "0 9 * * *",
    "priority": 3
  }'
```

### Webhook задача
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Notify Slack",
    "type": "send_webhook",
    "payload": "{\"url\":\"https://hooks.slack.com/services/xxx\",\"method\":\"POST\",\"headers\":{\"Content-Type\":\"application/json\"},\"body\":\"{\\\"text\\\":\\\"Deploy complete\\\"}\"}",
    "cron_expr": "*/30 * * * *",
    "priority": 1
  }'
```

### HTTP задача
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Health Check External API",
    "type": "http",
    "payload": "{\"url\":\"https://api.example.com/health\",\"method\":\"GET\",\"timeout\":10}",
    "cron_expr": "*/5 * * * *",
    "priority": 2
  }'
```

### Задача с ограниченным временем жизни
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Temporary Task",
    "type": "email",
    "payload": "{\"to\":\"admin@example.com\",\"subject\":\"Alert\",\"body\":\"Check system\"}",
    "cron_expr": "0 * * * *",
    "priority": 5,
    "max_retries": 5,
    "expires_at": "2026-04-01T00:00:00Z"
  }'
```

## Работа с приоритетами

Приоритеты задач от 1 (наивысший) до 10 (наименьший). По умолчанию — 5.

```bash
# Критическая задача (приоритет 1)
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Critical Alert",
    "type": "email",
    "payload": "{\"to\":\"oncall@example.com\",\"subject\":\"CRITICAL\",\"body\":\"System down\"}",
    "cron_expr": "* * * * *",
    "priority": 1
  }'

# Фоновая задача (приоритет 10)
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Weekly Report",
    "type": "email",
    "payload": "{\"to\":\"team@example.com\",\"subject\":\"Weekly\",\"body\":\"Report\"}",
    "cron_expr": "0 9 * * 1",
    "priority": 10
  }'
```

## Получение задачи по ID

```bash
curl http://localhost:8080/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000
```

## Dashboard API

### Статистика
```bash
curl http://localhost:8080/api/v1/dashboard/stats
```

Пример ответа:
```json
{
  "total_tasks": 150,
  "pending_tasks": 30,
  "scheduled_tasks": 50,
  "running_tasks": 5,
  "completed_tasks": 60,
  "failed_tasks": 5,
  "tasks_by_type": {"email": 80, "send_webhook": 50, "http": 20},
  "tasks_by_priority": {"1": 10, "3": 40, "5": 70, "10": 30}
}
```

### Список задач с фильтрацией
```bash
# Все задачи (пагинация по умолчанию: limit=20, offset=0)
curl http://localhost:8080/api/v1/dashboard/tasks

# Фильтр по статусу
curl "http://localhost:8080/api/v1/dashboard/tasks?status=failed"

# С пагинацией
curl "http://localhost:8080/api/v1/dashboard/tasks?limit=10&offset=20"

# Комбинация
curl "http://localhost:8080/api/v1/dashboard/tasks?status=pending&limit=5&offset=0"
```

## Мониторинг

### Health Check
```bash
curl http://localhost:8080/health
```

### Prometheus метрики
```bash
curl http://localhost:8080/metrics
```

### Grafana
Откройте http://localhost:3000 (admin/admin) после запуска:
```bash
docker-compose -f docker-compose.monitoring.yml up -d
```
