``` 
| 2_lab/
├── main.go
├── go.mod / go.sum
├── Dockerfile                  ← multi-stage build
├── docker-compose.yml          ← app + PostgreSQL + KeyDB
├── README.md                   ← ответы на все вопросы
└── internal/
├── app/server/app.go              ← инициализация и запуск
├── config/config.go        ← PGConfig + RedisConfig из env
├── handler/handler.go      ← GET/POST /notes, DELETE /notes/{id}, GET /health
├── repo/pg/repo.go         ← PostgreSQL CRUD + миграция
├── cache/redis/cache.go    ← KeyDB клиент, TTL=60s, ключ notes:all
└── usecase/usecase.go      ← cache-aside логика
```

## Как работает cache-aside:

```
Операция	Поведение
GET /notes	Проверяет notes:all в KeyDB → HIT: возвращает из кэша, MISS: идёт в БД, пишет в кэш
POST /notes	Пишет в БД → инвалидирует notes:all
DELETE /notes/{id}	Удаляет из БД → инвалидирует notes:all
```
## Запуск:

`docker-compose up --build`