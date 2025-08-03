# Wallet Service

REST‑сервис для управления балансами кошельков. Написан на **Go 1.23**, использует **PostgreSQL 14** и полностью контейнеризован.

---

## Быстрый старт

```bash
# 1. Клонируем репозиторий
 git clone https://github.com/Svshilkin/wallet‑service.git
 cd <...>/wallet‑service

# 2. Запускаем стек (база + API)
 docker-compose up --build -d

# 3. Проверяем логи
 docker-compose logs -f api
```

Cервис доступен на **http://localhost:8080**.

---

## Конфигурация

Все переменные перечислены в `config.env` ➊ и автоматически загружаются в приложение.

| Ключ            | Описание                    | Пример            |
|-----------------|-----------------------------|-------------------|
| `DB_HOST`       | адрес Postgres в контейнере | `db`              |
| `DB_PORT`       | порт Postgres в контейнере  | `5432`            |
| `DB_USER`       | логин                       | `wallet_user`     |
| `DB_PASSWORD`   | пароль                      | `wallet_password` |
| `DB_NAME`       | база                        | `wallet_db`       |
| `SERVER_PORT`   | внешний порт API            | `8080`            |

---

## API

### POST `/api/v1/wallet`

Изменяет баланс.
```json
{
  "valletId": "<UUID>",
  "operationType": "DEPOSIT" | "WITHDRAW",
  "amount": "<int>"
}
```
Ответ: `{ "balance": <int> }` или `409 Conflict`, если средств не хватает.

### GET `/api/v1/wallets/{walletId}`

Возвращает текущий баланс кошелька. 404, если id неизвестен.

---

## Тесты

```bash
# без детектора гонок
go test ./...

# с детектором гонок (gcc / clang обязателен)
CGO_ENABLED=1 go test -race ./...
```

---

## Команды

| Действие                      | Команда                                                   |
|-------------------------------|-----------------------------------------------------------|
| Пересобрать и запустить       | `docker-compose up --build -d`                            |
| Логи API                      | `docker-compose logs -f api`                              |
| Статус контейнеров            | `docker-compose ps`                                       |
| Остановка и удаление ресурсов | `docker-compose down -v`                                  |


