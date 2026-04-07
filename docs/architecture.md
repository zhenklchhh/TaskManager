# Архитектура TaskManager

## Обзор

TaskManager — распределённая система управления задачами, построенная на Go с использованием PostgreSQL и Redis.

## Компоненты системы

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   REST API      │     │     Worker       │     │   Scheduler     │
│   (Chi Router)  │     │  (Task Runner)   │     │   (Cron-based)  │
│                 │     │                  │     │                 │
│ - CRUD Tasks    │     │ - Dequeue tasks  │     │ - Cron parsing  │
│ - Dashboard     │     │ - Execute        │     │ - Auto schedule │
│ - Health checks │     │ - Retry logic    │     │ - Stale recovery│
│ - Metrics       │     │ - Error handling │     │                 │
└───────┬─────────┘     └────────┬─────────┘     └────────┬────────┘
        │                        │                         │
        ▼                        ▼                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Service Layer                              │
│  TaskService: CreateTask, RetryTask, ProcessPending, Stats      │
└───────────────────────────┬─────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
    ┌──────────────┐ ┌─────────────┐ ┌─────────────────┐
    │  PostgreSQL   │ │    Redis    │ │  Task Handlers  │
    │              │ │             │ │                 │
    │ - Tasks table│ │ - Queue     │ │ - EmailHandler  │
    │ - Migrations │ │ - Semaphore │ │ - WebhookHandler│
    │ - Stats      │ │ - Sorted    │ │ - GRPCHandler   │
    │              │ │   Sets      │ │ - HTTPHandler   │
    └──────────────┘ └─────────────┘ └─────────────────┘
```

## Потоки данных

### Создание задачи
1. Клиент → `POST /api/v1/tasks` → Handler
2. Handler → валидация → TaskService.CreateTask
3. TaskService → парсинг cron, установка defaults → Repository.Create
4. Repository → INSERT в PostgreSQL
5. Задача попадает в Redis Queue (Sorted Set по приоритету)

### Обработка задачи
1. Worker → Dequeue из Redis (по приоритету)
2. Семафор Redis → проверка слотов
3. Worker → TaskHandler.Execute (email/webhook/grpc)
4. Успех → UpdateTaskStatus(completed)
5. Ошибка → RetryTask (exponential backoff) или Failed

### Планирование
1. Scheduler → каждую минуту проверяет задачи с next_run_at <= now
2. Задачи со статусом scheduled → перемещает в pending
3. Stale задачи (running > threshold) → возврат в pending

## Пакетная структура

```
cmd/
├── api/main.go          # Точка входа API сервера
└── worker/worker.go     # Точка входа Worker

internal/
├── api/                 # HTTP handlers, routes, middleware
├── config/              # Конфигурация приложения
├── domain/              # Доменные сущности (Task, TaskStats, errors)
├── metrics/             # Prometheus метрики
├── queue/               # Redis очередь с приоритетами
├── repository/          # Интерфейс + PostgreSQL реализация
├── scheduler/           # Cron планировщик
├── service/             # Бизнес-логика (TaskService)
├── task/                # Обработчики задач (email, webhook, grpc, http)
└── worker/              # Worker для обработки задач

config/                  # YAML конфигурация
migrations/              # SQL миграции
monitoring/              # Prometheus + Grafana конфигурация
```

## Технологический стек

| Компонент | Технология |
|-----------|-----------|
| Язык | Go 1.21+ |
| HTTP Router | Chi v5 |
| База данных | PostgreSQL |
| Очередь | Redis (Sorted Sets) |
| Метрики | Prometheus |
| Визуализация | Grafana |
| Контейнеризация | Docker Compose |
| Миграции | golang-migrate |
| Валидация | go-playground/validator |
