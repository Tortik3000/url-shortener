# URL Shortener

## Запуск через `Makefile`:

| Команда         | Что делает                                                         |
|-----------------|--------------------------------------------------------------------|
| `make up`       | `docker compose up --build`    |
| `make down`     | `docker compose down`               |


## Примеры запросов

Готовые запросы для тестирования лежат в `tools/`:

| Файл                                  | Что внутри                                                      |
|---------------------------------------|-----------------------------------------------------------------|
| `tools/shorten.http`                  | `POST /shorten` — happy-path, идемпотентность, все ветки ошибок |
| `tools/resolve.http`                  | `GET /{code}` — 302-редирект, `NOT_FOUND`, `INVALID_CODE`       |
| `tools/http-client.environment.json`  | Окружения `local`  (переменные `baseUrl`, `sampleURL`)          |

## API

### `POST /shorten`

Создаёт сокращённую ссылку. Идемпотентен: повторный запрос с тем же URL возвращает тот же код.

**Запрос:**
```json
{"url": "https://example.com/very/long/path"}
```

**Ответ 201 Created:**
```json
{
  "short_code": "aB3_xY9Zk_",
  "short_url": "http://localhost:8080/aB3_xY9Zk_"
}
```

**Ошибки:** `400 INVALID_URL`, `400 INVALID_REQUEST`, `503 RETRY_LATER`, `500 INTERNAL_ERROR`.

### `GET /{code}`

Возвращает `302 Found` с заголовком `Location: <original_url>`. 

**Ошибки:** `400 INVALID_CODE`, `404 NOT_FOUND`, `500 INTERNAL_ERROR`.

## Переменные окружения

| Переменная             | По умолчанию             | Описание                                                                       |
|------------------------|--------------------------|--------------------------------------------------------------------------------|
| `PORT`                 | `8080`                   | Порт HTTP-сервера                                                              |
| `STORAGE`              | `postgres`               | `postgres` или `memory`                                                        |
| `DATABASE_URL`         | —                        | Обязателен при `STORAGE=postgres` (формат pgx)                                 |
| `BASE_URL`             | `http://localhost:8080`  | База для поля `short_url` в ответе POST /shorten                               |
| `SNOWFLAKE_MACHINE_ID` | `0`                      | Уникальный ID инстанса для Snowflake-генератора (0..255)                       |
