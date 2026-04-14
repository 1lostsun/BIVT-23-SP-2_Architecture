```
3_lab/
├── docker-compose.yml          ← 4 контейнера
├── README.md                   ← ответы на все 9 вопросов
└── services/
    ├── notes-api/              ← Producer (REST API)
    │   ├── Dockerfile
    │   ├── internal/
    │   │   ├── app/            ← инициализация
    │   │   ├── config/         ← PGConfig + RabbitConfig из env
    │   │   ├── handler/        ← GET/POST/DELETE /notes
    │   │   ├── publisher/      ← публикация событий в RabbitMQ
    │   │   ├── repo/pg/        ← PostgreSQL CRUD
    │   │   └── usecase/        ← бизнес-логика + вызов publisher
    └── notifier/               ← Consumer
        ├── Dockerfile
        └── main.go             ← подписка на очередь, логирование
```

## Event driven flow 

```
POST /notes
   └─→ PostgreSQL (INSERT)
   └─→ RabbitMQ Exchange notes.events
           topic routingKey: note.created
               └─→ Queue notifier.queue  (binding: note.*)
                       └─→ notifier logs: [notifier] received event: type=note.created ...
```

## Запуск 

```
docker-compose up --build
# RabbitMQ UI → http://localhost:15672  (guest/guest)
# API         → http://localhost:8080
 ```

