## Конфигурация

Основные параметры конфигурации в `config/local.yaml`:

- `server.address` - адрес HTTP сервера
- `db.url` - строка подключения к PostgreSQL
- `redis.address` - адрес Redis
- `scheduler.stale-task-threshold` - порог для определения "застарелых" задач
- `default-max-retries` - максимальное количество повторных попыток

## Environment Variables

- `CONFIG_PATH` - путь к файлу конфигурации
- `DATABASE_URL` - URL подключения к БД
- `REDIS_ADDR` - адрес Redis сервера
- `STALE_TASK_THRESHOLD` - порог для застарелых задач