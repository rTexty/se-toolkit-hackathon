[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/uvnTmvcw)

# Room Booking Service

Сервис бронирования переговорок — тестовое задание.

## Запуск

```bash
make up
make seed
make down
```

Сервис на `http://localhost:8080`.

## Тесты

```bash
make test
make test-integration
make test-load
```

## Swagger

Документация доступна по адресу `http://localhost:8080/docs/`.

Сгенерирована через `swag init`.

## Архитектурные решения

### Генерация слотов

Слоты генерируются on-demand при запросе `/slots/list?date=X`.
UUID детерминированный (SHA-256 от roomID + startTime) — стабильные ID.
Слоты сохраняются через upsert, повторные запросы не дублируют данные.

### Почему on-demand

50 переговорок x ~12 слотов/день = ~600 слотов/день.
Pre-generation на 7 дней — ~4200 записей, большинство не запрошены.
On-demand генерирует только то, что нужно, и кеширует в БД.

### Авторизация

JWT с `user_id` и `role`. Фиксированные UUID для dummyLogin:

- admin: `00000000-0000-0000-0000-000000000001`
- user: `00000000-0000-0000-0000-000000000002`

### Регистрация и логин

Эндпоинты `/register` и `/login` реализованы как дополнительное задание.
Пароли хешируются через bcrypt. При регистрации возвращается JWT в заголовке `X-Auth-Token`.

### Conference Link

При передаче `createConferenceLink: true` в запросе создания брони генерируется мок-ссылка на конференцию.
Реальный внешний сервис не вызывается — ссылка генерируется локально.
При сбое внешнего сервиса бронь всё равно создаётся, но без conferenceLink.

## Дополнительные задания

| Задание | Статус |
|---------|--------|
| /register и /login | ✅ |
| createConferenceLink | ✅ |
| Makefile | ✅ |
| Swagger-документация | ✅ |
| Нагрузочное тестирование | ✅ (см. tests/load_test.go) |
| .golangci.yaml | ✅ |
