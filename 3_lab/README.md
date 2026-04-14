# Лабораторная работа №3 — Event-Driven Architecture с RabbitMQ

## Архитектура проекта

Проект состоит из двух Go-микросервисов, взаимодействующих через брокер сообщений RabbitMQ:

- **notes-api** — REST API (Producer): принимает HTTP-запросы, сохраняет заметки в PostgreSQL и публикует события в RabbitMQ.
- **notifier** — Consumer: подписывается на очередь RabbitMQ и логирует полученные события.

```
┌────────────┐   HTTP    ┌───────────┐   AMQP    ┌──────────┐   AMQP    ┌──────────┐
│   Client   │ ────────► │ notes-api │ ────────► │ RabbitMQ │ ────────► │ notifier │
└────────────┘           └───────────┘           └──────────┘           └──────────┘
                               │                      
                               │ SQL                  
                               ▼                      
                         ┌───────────┐               
                         │ PostgreSQL│               
                         └───────────┘               
```

---

## Запуск

```bash
docker-compose up --build
```

После запуска:
- **RabbitMQ Management UI**: http://localhost:15672 (guest / guest)
- **REST API**: http://localhost:8080

### Примеры запросов

```bash
# Проверка здоровья сервиса
curl http://localhost:8080/health

# Получить все заметки
curl http://localhost:8080/notes

# Создать заметку
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","body":"Hello RabbitMQ"}'

# Удалить заметку по ID
curl -X DELETE http://localhost:8080/notes/1
```

При создании или удалении заметки в логах контейнера **notifier** появится сообщение о полученном событии.

---

## Структура каталогов

```
3_lab/
├── docker-compose.yml
├── README.md
└── services/
    ├── notes-api/
    │   ├── Dockerfile
    │   ├── go.mod
    │   ├── go.sum
    │   ├── main.go
    │   └── internal/
    │       ├── app/app.go          — инициализация и запуск приложения
    │       ├── config/config.go    — конфигурация из переменных окружения
    │       ├── handler/handler.go  — HTTP-обработчики (REST API)
    │       ├── repo/pg/repo.go     — репозиторий (PostgreSQL)
    │       ├── publisher/publisher.go — публикация событий в RabbitMQ
    │       └── usecase/usecase.go  — бизнес-логика
    └── notifier/
        ├── Dockerfile
        ├── go.mod
        ├── go.sum
        └── main.go                — подписка на очередь и логирование событий
```

---

## Ответы на вопросы

### 1. Что такое Event-Driven Architecture, какие ключевые компоненты она включает?

**Event-Driven Architecture (EDA)** — это архитектурный стиль, при котором компоненты системы взаимодействуют посредством событий (events), а не прямых вызовов друг друга. Вместо того чтобы сервис A синхронно вызывал API сервиса B, сервис A публикует событие в брокер, а сервис B (или любое количество других сервисов) реагирует на него асинхронно.

**Отличие от Request-Driven архитектуры:**

| Критерий | Request-Driven (REST/gRPC) | Event-Driven |
|---|---|---|
| Инициатор | Потребитель (Consumer) делает запрос | Производитель (Producer) публикует событие |
| Связанность | Сильная (Producer знает о Consumer) | Слабая (Producer не знает о Consumer) |
| Синхронность | Синхронная (ждёт ответа) | Асинхронная (fire-and-forget) |
| Масштабирование | Consumer должен быть доступен | Consumer может быть временно недоступен |

**Ключевые компоненты EDA:**

1. **Event Producer (Производитель)** — компонент, генерирующий события при изменении состояния системы. В нашем проекте: сервис `notes-api`, который публикует события `note.created` и `note.deleted` после записи в БД.

2. **Event Broker (Брокер событий)** — посредник, принимающий события от производителей и маршрутизирующий их потребителям. В нашем проекте: **RabbitMQ**. Брокер обеспечивает надёжную доставку, буферизацию и маршрутизацию сообщений.

3. **Event Consumer (Потребитель)** — компонент, подписывающийся на события и реагирующий на них. В нашем проекте: сервис `notifier`, который получает события и логирует их. Потребителей может быть несколько — каждый обрабатывает событие независимо.

4. **Event (Событие)** — само сообщение, описывающее факт произошедшего изменения. В нашем проекте событие содержит тип (`note.created`/`note.deleted`), полезную нагрузку (данные заметки или её ID) и временную метку.

---

### 2. Какие есть плюсы и минусы от использования Event-Driven взаимодействия?

#### Плюсы

**Слабая связанность (Loose Coupling):**
Producer не знает о существовании Consumer'ов. Можно добавлять новые сервисы-подписчики, не изменяя Producer. Например, можно добавить сервис отправки email-уведомлений, просто подписав его на ту же очередь — `notes-api` не нужно переписывать.

**Масштабируемость (Scalability):**
Consumer'ов можно масштабировать горизонтально независимо от Producer'ов. Несколько экземпляров `notifier` могут читать из одной очереди, распределяя нагрузку. Broker выступает буфером при пиковых нагрузках.

**Асинхронность (Asynchronous Processing):**
Producer не ждёт ответа от Consumer'а — он продолжает работу немедленно после публикации события. Это сокращает время ответа API и повышает производительность.

**Отказоустойчивость (Resilience):**
Если Consumer временно недоступен, сообщения накапливаются в очереди брокера. После восстановления Consumer обработает все накопленные события. В нашем проекте `notifier` имеет retry-логику подключения к RabbitMQ.

**Возможность повторного воспроизведения (Replay):**
При наличии журнала событий (event log) можно воспроизвести историю событий для восстановления состояния или отладки.

#### Минусы

**Сложность отладки (Debugging Complexity):**
Асинхронная природа усложняет трассировку: трудно понять, какой Consumer обработал какое событие и когда. Требуются инструменты распределённой трассировки (Jaeger, Zipkin) и structured logging.

**Eventual Consistency (Конечная согласованность):**
Данные между сервисами согласованы не мгновенно, а через некоторое время после обработки события. Это усложняет проектирование бизнес-логики, где требуется немедленная согласованность.

**Дублирование сообщений (Duplicate Messages):**
Сбои при подтверждении (ACK) могут привести к повторной доставке одного и того же события. Consumer'ы должны быть идемпотентными — повторная обработка одного события не должна менять конечный результат.

**Сложность транзакций (Distributed Transactions):**
Невозможно атомарно записать в БД и опубликовать событие стандартными средствами. Если запись в БД прошла успешно, а публикация события — нет (или наоборот), система приходит в несогласованное состояние. Решение: паттерн Transactional Outbox.

**Операционная сложность:**
Необходимо поддерживать брокер сообщений как отдельный компонент инфраструктуры: мониторинг, резервирование, управление схемами сообщений (schema registry).

---

### 3. Для чего нужен и какую роль играет RabbitMQ? Какой протокол он реализует?

**RabbitMQ** — это брокер сообщений (message broker) с открытым исходным кодом, написанный на Erlang. Он выступает посредником между сервисами, принимая сообщения от Producer'ов и надёжно доставляя их Consumer'ам.

**Роль в системе:**
- **Буферизация:** принимает сообщения и хранит их в очередях до тех пор, пока Consumer не будет готов их обработать.
- **Маршрутизация:** на основе типа exchange и routing key определяет, в какую очередь (или очереди) направить сообщение.
- **Надёжность доставки:** поддерживает подтверждения (ACK/NACK), персистентные сообщения и очереди для обеспечения гарантий доставки.
- **Декаплинг:** позволяет Producer'ам и Consumer'ам работать независимо, не зная ничего друг о друге.
- **Масштабирование:** поддерживает кластеризацию и зеркалирование очередей для высокой доступности.

**Протокол:**
RabbitMQ реализует протокол **AMQP 0-9-1** (Advanced Message Queuing Protocol). Это открытый стандарт для обмена сообщениями, определяющий формат фреймов, модель обмена (Producer → Exchange → Queue → Consumer), механизмы подтверждения и управления потоком. В нашем проекте используется Go-библиотека `github.com/rabbitmq/amqp091-go`, реализующая клиент этого протокола.

Помимо AMQP 0-9-1, RabbitMQ поддерживает MQTT, STOMP и HTTP через плагины.

---

### 4. Какие основные сущности есть в RabbitMQ? Как выполняется коммуникация между Producer и Consumer?

**Основные сущности RabbitMQ:**

- **Connection** — TCP-соединение между клиентом (Producer или Consumer) и RabbitMQ-сервером. Создание соединения — дорогостоящая операция.

- **Channel** — виртуальный канал внутри Connection. Один Connection может содержать множество Channel'ов. Большинство операций AMQP выполняется через Channel, что позволяет мультиплексировать соединение.

- **Virtual Host (vhost)** — логическое разделение брокера, аналог базы данных в PostgreSQL. Каждый vhost имеет собственные Exchange'и, Queue'и и права доступа. По умолчанию используется vhost `/`.

- **Exchange** — точка входа для сообщений от Producer'а. Exchange принимает сообщение и на основе его routing key и правил (bindings) направляет его в одну или несколько очередей. Exchange сам по себе не хранит сообщения.

- **Queue** — буфер (очередь), в котором сообщения хранятся до тех пор, пока Consumer не получит и не подтвердит их. Очереди могут быть durable (переживают перезапуск RabbitMQ) или transient.

- **Binding** — правило, связывающее Exchange с Queue. При создании binding задаётся binding key (паттерн routing key), по которому Exchange решает, направлять ли сообщение в данную Queue.

- **Routing Key** — строка-метка, которую Producer указывает при публикации сообщения. Exchange использует её для маршрутизации.

**Схема коммуникации Producer → Consumer:**

```
Producer
  │
  │ channel.Publish(exchange="notes.events", routingKey="note.created", body)
  ▼
Exchange (notes.events, topic)
  │
  │ Binding: routingKey pattern "note.*" → Queue "notifier.queue"
  ▼
Queue (notifier.queue)
  │
  │ channel.Consume(queue="notifier.queue")
  ▼
Consumer (notifier)
  │
  │ msg.Ack() — подтверждение обработки
  ▼
  (сообщение удаляется из очереди)
```

1. Producer подключается к RabbitMQ, создаёт Channel и объявляет Exchange.
2. Consumer подключается, создаёт Channel, объявляет Exchange и Queue, создаёт Binding.
3. Producer публикует сообщение в Exchange с указанием routing key.
4. Exchange находит все Queue, связанные через Binding с подходящим pattern, и направляет туда копию сообщения.
5. Consumer читает сообщение из Queue и после обработки отправляет ACK.
6. RabbitMQ удаляет сообщение из Queue после получения ACK.

---

### 5. Из каких частей состоит сообщение в RabbitMQ? Что такое routing key сообщения?

**Сообщение в RabbitMQ состоит из двух частей:**

**1. Properties (метаданные / заголовки):**
Структурированные атрибуты сообщения, определённые протоколом AMQP:

| Свойство | Описание | Пример в проекте |
|---|---|---|
| `content-type` | MIME-тип тела сообщения | `application/json` |
| `delivery-mode` | 1 — transient, 2 — persistent (переживает рестарт) | `2` (Persistent) |
| `priority` | Приоритет сообщения (0–9) | не используется |
| `correlation-id` | ID для связи запроса и ответа (паттерн RPC) | не используется |
| `reply-to` | Имя очереди для ответа (паттерн RPC) | не используется |
| `expiration` | TTL сообщения в миллисекундах | не используется |
| `message-id` | Уникальный идентификатор сообщения | не используется |
| `timestamp` | Unix timestamp создания сообщения | не используется |
| `type` | Тип сообщения (произвольная строка) | не используется |
| `headers` | Произвольные key-value заголовки | не используется |

**2. Body (тело):**
Произвольный набор байт — полезная нагрузка сообщения. RabbitMQ не интерпретирует тело; его формат определяется соглашением между Producer и Consumer.

В нашем проекте тело — JSON-объект следующей структуры:
```json
{
  "type": "note.created",
  "payload": {"id": 1, "title": "Test", "body": "Hello RabbitMQ"},
  "timestamp": "2024-04-14T12:00:00Z"
}
```

**Routing Key:**
Routing key — это строка, которую Producer указывает при публикации сообщения (метаданные на уровне протокола, отдельные от заголовков сообщения). Exchange использует routing key для принятия решения о маршрутизации: в какую Queue (или Queue'и) направить данное сообщение.

Routing key не является частью тела или заголовков сообщения — это параметр команды `basic.publish`. Правила интерпретации routing key зависят от типа Exchange:
- **Direct**: routing key должен точно совпадать с binding key.
- **Topic**: routing key проверяется на соответствие паттерну с wildcards.
- **Fanout**: routing key игнорируется полностью.
- **Headers**: routing key игнорируется, используются заголовки сообщения.

В нашем проекте routing key'и: `note.created` и `note.deleted`.

---

### 6. Какие виды Exchange поддерживает RabbitMQ?

RabbitMQ поддерживает четыре встроенных типа Exchange:

**1. Direct Exchange**
Маршрутизирует сообщение в Queue, если routing key сообщения **точно совпадает** с binding key привязки. Используется для точной адресации.

```
Producer → routingKey="error"
Exchange (direct) → binding key "error" → Queue "error-queue"
                  → binding key "info"  → (не совпадает, игнорируется)
```

**2. Fanout Exchange**
Маршрутизирует сообщение **во все** Queue, привязанные к данному Exchange. Routing key полностью игнорируется. Аналог широковещательной рассылки (broadcast). Используется, когда одно событие должны получить все подписчики.

```
Producer → routingKey="anything" (игнорируется)
Exchange (fanout) → Queue "service-a"
                  → Queue "service-b"
                  → Queue "service-c"
```

**3. Topic Exchange** *(используется в нашем проекте)*
Маршрутизирует сообщение на основе **паттерн-матчинга** routing key. Routing key — строка слов, разделённых точкой (`word1.word2.word3`). Binding key может содержать wildcards:
- `*` — ровно одно слово
- `#` — ноль или более слов

```
Producer → routingKey="note.created"
Exchange (topic) → binding "note.*"    → Queue "notifier.queue" (совпадает)
                 → binding "note.#"    → Queue "audit.queue" (совпадает)
                 → binding "order.*"   → (не совпадает)
```

В нашем проекте: Exchange `notes.events` типа `topic`. Consumer `notifier` подписан с паттерном `note.*`, что позволяет получать как `note.created`, так и `note.deleted`.

**4. Headers Exchange**
Маршрутизирует сообщение на основе **заголовков** (properties.headers) сообщения, а не routing key. Binding задаёт набор заголовков и условие их проверки (`x-match: all` — все должны совпасть, `x-match: any` — хотя бы один). Routing key игнорируется. Используется редко — когда маршрутизация на основе routing key недостаточно гибка.

```
Producer → headers: {format: "pdf", type: "report"}
Exchange (headers) → binding {format: "pdf", x-match: "any"} → Queue "pdf-queue" (совпадает)
```

---

### 7. Чем direct exchange отличается от topic exchange?

| Характеристика | Direct Exchange | Topic Exchange |
|---|---|---|
| Тип совпадения | Точное равенство строк | Паттерн-матчинг с wildcards |
| Wildcards | Не поддерживаются | `*` (одно слово), `#` (ноль или более слов) |
| Гибкость маршрутизации | Низкая — один routing key → одна Queue | Высокая — один паттерн → множество routing key'ев |
| Типичное применение | Балансировка нагрузки, точная адресация | Иерархические события, категоризация сообщений |
| Производительность | Чуть выше (точное совпадение) | Чуть ниже (regexp-матчинг) |

**Пример:**

*Direct Exchange:*
```
Binding key: "note.created"
Routing key "note.created" → совпадает
Routing key "note.deleted" → НЕ совпадает (нужен отдельный binding)
Routing key "note.updated" → НЕ совпадает (нужен отдельный binding)
```

*Topic Exchange:*
```
Binding key: "note.*"
Routing key "note.created" → совпадает (* = "created")
Routing key "note.deleted" → совпадает (* = "deleted")
Routing key "note.updated" → совпадает (* = "updated")
Routing key "order.created" → НЕ совпадает

Binding key: "note.#"
Routing key "note.created"        → совпадает
Routing key "note.created.batch"  → совпадает (# = "created.batch")
Routing key "note"                → совпадает (# = "")
```

**Когда использовать:**
- **Direct**: когда каждый тип события должен попадать в строго определённую очередь без обобщений. Например, логирование уровней `debug`, `info`, `warning`, `error` в разные очереди.
- **Topic**: когда нужна гибкая подписка на категории событий. Например, аудит-сервис подписывается на `note.#` (все события с заметками), а notifier — только на `note.created`.

---

### 8. Как выполняется связь между exchange и queue?

Связь между Exchange и Queue устанавливается через **Binding** (привязку).

**Binding** — это правило маршрутизации, которое говорит Exchange: «при получении сообщения с routing key, соответствующим данному паттерну, направь его в указанную Queue».

**Создание Binding в коде:**
```go
err = ch.QueueBind(
    q.Name,      // имя очереди
    routingKey,  // binding key (паттерн): "note.*"
    exchange,    // имя exchange: "notes.events"
    false,       // noWait
    nil,         // аргументы
)
```

**Ключевые особенности:**

1. **Один Exchange — несколько Queue:** один Exchange может быть связан с множеством Queue через разные binding key. Это позволяет маршрутизировать разные типы событий в разные очереди.

2. **Одна Queue — несколько Exchange:** одна Queue может принимать сообщения от нескольких Exchange'ей через разные Binding'и.

3. **Несколько Binding'ов между одним Exchange и одной Queue:** можно создать несколько Binding с разными routing key паттернами. Если хотя бы один совпал, сообщение попадёт в Queue (но только один раз).

4. **Для Fanout Exchange** binding key не имеет значения — Exchange направляет сообщение во все привязанные Queue вне зависимости от binding key.

5. **Default Exchange** — безымянный встроенный Direct Exchange. К нему автоматически привязывается каждая Queue с binding key, равным имени Queue. Позволяет публиковать напрямую в Queue по имени без явного создания Binding.

**Жизненный цикл Binding:**
Binding существует до тех пор, пока не будет явно удалён (`QueueUnbind`) или пока Queue/Exchange не будут удалены.

---

### 9. Для чего и как используется RabbitMQ в вашем приложении?

**Цель использования:**
RabbitMQ используется для реализации асинхронного уведомления об изменениях в системе заметок. Сервис `notes-api` не знает о существовании `notifier` — он просто публикует факт произошедшего события, а `notifier` независимо реагирует на него. Это позволяет добавлять новые реакции на события (email-уведомления, аудит, аналитика) без изменения `notes-api`.

**Конфигурация RabbitMQ в проекте:**
- **Exchange**: `notes.events`, тип `topic`, durable (персистентный)
- **Queue**: `notifier.queue`, durable
- **Binding**: паттерн `note.*` связывает Exchange с Queue
- **Routing key'и**: `note.created`, `note.deleted`

**Полный flow создания заметки:**

```
1. Client: POST /notes {"title":"Test","body":"Hello RabbitMQ"}
          │
          ▼
2. notes-api / handler: декодирует JSON, вызывает usecase.CreateNote()
          │
          ▼
3. usecase.CreateNote():
   a) repo.Create() → INSERT INTO notes ... RETURNING id, title, body
          │
          ▼
   b) publisher.Publish("note.created", note)
          │
          ▼
4. publisher.Publish():
   - Сериализует payload: {"id":1,"title":"Test","body":"Hello RabbitMQ"}
   - Создаёт Event: {"type":"note.created","payload":{...},"timestamp":"..."}
   - channel.PublishWithContext(exchange="notes.events", routingKey="note.created", body)
          │
          ▼
5. RabbitMQ Exchange "notes.events" (topic):
   - routing key "note.created" соответствует binding pattern "note.*"
   - Копирует сообщение в Queue "notifier.queue"
          │
          ▼
6. notifier / Consumer:
   - Получает сообщение из "notifier.queue"
   - Десериализует JSON → Event{Type:"note.created", Payload:{...}}
   - Логирует: [notifier] received event: type=note.created payload={"id":1,...} timestamp=...
   - msg.Ack(false) — подтверждает обработку, RabbitMQ удаляет из очереди
          │
          ▼
7. Client: получает ответ 201 Created {"id":1,"title":"Test","body":"Hello RabbitMQ"}
```

**Flow удаления заметки:**

```
Client: DELETE /notes/1
  → usecase.DeleteNote(1)
  → repo.Delete(1)           → DELETE FROM notes WHERE id=1
  → publisher.Publish("note.deleted", {"id":1})
  → RabbitMQ → notifier.queue
  → notifier: [notifier] received event: type=note.deleted payload={"id":1} timestamp=...
```

**Гарантии надёжности:**
- Exchange и Queue объявлены как `durable=true` — переживают перезапуск RabbitMQ.
- Сообщения публикуются с `DeliveryMode: amqp.Persistent` — тело сообщения сохраняется на диск.
- Consumer использует ручное подтверждение (`autoAck=false`) — если `notifier` упадёт в процессе обработки, сообщение вернётся в очередь и будет доставлено снова.
- `notifier` имеет retry-цикл (10 попыток по 3 секунды) для подключения к RabbitMQ при старте.
