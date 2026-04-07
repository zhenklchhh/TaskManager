# Troubleshooting

## Частые проблемы и решения

### Подключение к PostgreSQL

**Симптом:** `connection refused` при запуске сервера

**Решение:**
1. Убедитесь, что PostgreSQL запущен:
   ```bash
   docker-compose ps
   ```
2. Проверьте параметры подключения в `.env` или `config/local.yaml`
3. Убедитесь, что порт 5432 не занят другим процессом:
   ```bash
   netstat -tlnp | grep 5432
   ```

### Подключение к Redis

**Симптом:** `dial tcp: connection refused` для Redis

**Решение:**
1. Проверьте, что Redis запущен:
   ```bash
   docker-compose ps
   redis-cli ping
   ```
2. Проверьте порт (по умолчанию 6379) в конфигурации
3. Убедитесь, что Redis доступен:
   ```bash
   redis-cli -h localhost -p 6379 ping
   ```

### Миграции не применяются

**Симптом:** `relation "tasks" does not exist`

**Решение:**
```bash
make migrate-up
```
Если ошибка persists:
1. Проверьте DATABASE_URL в `.env`
2. Убедитесь, что БД создана
3. Попробуйте сбросить и применить заново:
   ```bash
   make migrate-down
   make migrate-up
   ```

### Задачи не обрабатываются

**Симптом:** Задачи остаются в статусе `pending`

**Решение:**
1. Убедитесь, что Worker запущен:
   ```bash
   make run-worker
   ```
2. Проверьте подключение Worker к Redis
3. Проверьте семафор — возможно, все слоты заняты:
   ```bash
   curl http://localhost:8080/metrics | grep semaphore
   ```
4. Проверьте, что `next_run_at` задачи в прошлом

### Задачи переходят в failed

**Симптом:** Задачи быстро переходят в `failed` после retry

**Решение:**
1. Проверьте логи Worker для подробностей ошибки
2. Убедитесь, что payload задачи корректный
3. Для webhook задач — проверьте доступность целевого URL
4. Увеличьте `max_retries` при создании задачи

### Prometheus не собирает метрики

**Симптом:** Grafana показывает "No data"

**Решение:**
1. Проверьте endpoint метрик:
   ```bash
   curl http://localhost:8080/metrics
   ```
2. Убедитесь, что Prometheus настроен на правильный target в `monitoring/prometheus.yml`
3. Проверьте статус target в Prometheus UI: http://localhost:9090/targets

### Grafana dashboard не загружается

**Симптом:** Dashboard пустой после импорта

**Решение:**
1. Убедитесь, что datasource Prometheus настроен
2. Проверьте URL Prometheus в datasource settings
3. Импортируйте dashboard из `monitoring/grafana/dashboards/taskmanager.json`

### Порт уже занят

**Симптом:** `bind: address already in use`

**Решение:**
```bash
# Найти процесс на порту 8080
lsof -i :8080
# или на Windows
netstat -ano | findstr :8080

# Убить процесс или изменить порт в конфигурации
```

## Отладка задач

### Просмотр логов
Логи выводятся в stdout в формате slog. Уровень логирования настраивается в конфигурации.

### Проверка статуса через Dashboard API
```bash
# Посмотреть failed задачи
curl "http://localhost:8080/api/v1/dashboard/tasks?status=failed"

# Общая статистика
curl http://localhost:8080/api/v1/dashboard/stats
```

### Проверка Redis очереди
```bash
redis-cli ZCARD task_queue           # Количество задач в очереди
redis-cli ZRANGE task_queue 0 -1     # Все задачи в очереди
```
