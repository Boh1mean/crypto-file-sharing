# CryptoCore

CryptoCore — это прототип системы защищённой передачи файлов с end-to-end шифрованием, написанный на Go.
Проект построен вокруг следующих принципов:
- приватные ключи остаются на стороне клиента;
- сервер хранит только публичные ключи, зашифрованные контейнеры и метаданные файлов;
- шифрование, подписание, проверка подписи и расшифровка происходят локально в desktop-клиенте.

На текущем этапе проект уже включает:
- криптографическое ядро для сборки и расшифровки защищённых контейнеров;
- HTTP-сервер для хранения пользователей и зашифрованных контейнеров;
- desktop-клиент на `Fyne` для регистрации, отправки, получения и расшифровки файлов;
- автоматические тесты для криптографического ядра, сервисов и HTTP-транспорта.

## Текущая Архитектура

Проект разделён на несколько слоёв.

`internal/core`
- `crypto`: низкоуровневые криптографические примитивы, такие как `ECDH`, `ECDSA`, `AES-GCM`, `HKDF`, `SHA-256`
- `container`: протокол контейнера для файла, включая шифрование, обёртку file key, подпись, проверку подписи и расшифровку

`internal/domain`
- доменные модели и DTO для серверных операций с пользователями и файлами

`internal/repository`
- интерфейсы репозиториев пользователей, метаданных файлов и хранилища контейнеров

`internal/infrastructure/memory`
- in-memory реализации репозиториев для разработки и тестирования

`internal/service`
- серверная бизнес-логика
- `UserService`: создание пользователя и получение публичных ключей
- `FileService`: сохранение и загрузка зашифрованных контейнеров

`internal/transport`
- HTTP transport-слой
- маршруты, DTO запросов/ответов и обработчики

`internal/client`
- логика desktop-клиента
- `api`: HTTP-клиент для взаимодействия с сервером
- `keystore`: локальное хранение профиля и приватных ключей
- `app`: клиентские use case’ы, такие как регистрация, отправка и получение
- `ui`: desktop-интерфейс на `Fyne`

`cmd`
- `cmd/desktop`: точка входа desktop-приложения
- `cmd/devkeygen`: вспомогательная утилита для генерации тестового payload для `POST /users`

## Что Было Реализовано

### 1. Криптографическое Ядро

Реализовано в:
- [internal/core/crypto/crypto.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/core/crypto/crypto.go)
- [internal/core/container/container.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/core/container/container.go)

Доступная функциональность:
- генерация пар ключей `ECDH` и `ECDSA`
- шифрование и расшифровка через `AES-GCM`
- вывод ключей через `HKDF`
- хэширование через `SHA-256`
- сборка защищённого контейнера для файла
- проверка подписи контейнера
- расшифровка контейнера и проверка целостности

### 2. Серверная Модель Хранения

Сервер следует E2E-friendly модели:
- хранит только публичные ключи пользователей;
- хранит только зашифрованные контейнеры файлов и их метаданные;
- не хранит файлы в открытом виде;
- не использует приватные пользовательские ключи.

### 3. Слой Репозиториев

Реализованы интерфейсы:
- `UserRepository`
- `FileRepository`
- `ContainerStorage`

Реализованы in-memory версии:
- [internal/infrastructure/memory/user_repo.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/infrastructure/memory/user_repo.go)
- [internal/infrastructure/memory/file_repo.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/infrastructure/memory/file_repo.go)
- [internal/infrastructure/memory/container_repo.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/infrastructure/memory/container_repo.go)

### 4. Серверные Сервисы

Реализованы в:
- [internal/service/user_service.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/service/user_service.go)
- [internal/service/file_service.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/service/file_service.go)

Доступная функциональность:
- регистрация пользователя с публичными ключами
- получение публичных ключей по `userID`
- сохранение зашифрованного контейнера и метаданных файла
- загрузка зашифрованного контейнера и метаданных файла

### 5. HTTP API

Реализовано в:
- [internal/transport/httpapi/handler.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/transport/httpapi/handler.go)
- [internal/transport/httpapi/router.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/transport/httpapi/router.go)
- [internal/transport/httpapi/dto.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/transport/httpapi/dto.go)
- [internal/transport/router.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/transport/router.go)

Доступные endpoints:
- `POST /users`
- `GET /users/{id}/public-keys`
- `POST /files`
- `GET /files/{id}`

### 6. Точка Входа Сервера

Реализована в:
- [main.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/main.go)

Сервер запускается на:
- `http://localhost:8080`

### 7. Автоматические Тесты

Реализованы и проходят:
- тесты криптографии
- тесты контейнера
- тесты сервисов
- тесты HTTP transport-слоя

Основные тестовые файлы:
- [internal/core/crypto/crypto_test.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/core/crypto/crypto_test.go)
- [internal/core/container/container_test.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/core/container/container_test.go)
- [internal/service/file_service_test.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/service/file_service_test.go)
- [internal/service/user_service_test.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/service/user_service_test.go)
- [internal/transport/httpapi/handler_test.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/transport/httpapi/handler_test.go)

### 8. Desktop-Клиент На Fyne

Реализован в:
- [cmd/desktop/main.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/cmd/desktop/main.go)
- [internal/client/ui/app.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/client/ui/app.go)

Текущий интерфейс включает вкладки:
- `Register`
- `Send`
- `Receive`

### 9. Локальное Хранилище Ключей Клиента

Реализовано в:
- [internal/client/keystore/store.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/client/keystore/store.go)
- [internal/client/keystore/profile.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/client/keystore/profile.go)

Доступная функциональность:
- сохранение локального профиля
- загрузка локального профиля
- восстановление приватных ключей из профиля

### 10. Клиентские Use Case’ы

Реализованы в:
- [internal/client/app/register_user.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/client/app/register_user.go)
- [internal/client/app/send_file.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/client/app/send_file.go)
- [internal/client/app/receive_file.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/internal/client/app/receive_file.go)

Клиент на текущем этапе умеет:
- локально генерировать ключи и регистрировать пользователя
- читать локальный файл и собирать зашифрованный контейнер
- загружать зашифрованный контейнер на сервер
- загружать зашифрованный контейнер с сервера
- проверять подпись отправителя
- локально расшифровывать файл
- сохранять расшифрованный файл на диск

### 11. Вспомогательная Dev-Утилита

Реализована в:
- [cmd/devkeygen/main.go](/Users/danilknyazkin/Documents/ИВТ/Диплом/Мой Диплом/CrytoCore/cmd/devkeygen/main.go)

Назначение:
- генерация JSON payload для `POST /users`
- упрощение тестирования через Postman или curl

## Что Доступно Уже Сейчас

На текущем этапе проект уже поддерживает следующий практический сценарий:
- запуск сервера;
- регистрация пользователя через desktop-клиент;
- создание второго пользователя через Postman или helper-утилиту;
- отправка файла через desktop-клиент;
- проверка того, что контейнер сохранён на сервере;
- получение и расшифровка файла через desktop-клиент.

То есть проект уже содержит базовый рабочий E2E-прототип.

## Как Запустить Проект

### 1. Запуск Тестов

```bash
env GOCACHE=/tmp/codex-gocache go test ./...
```

### 2. Запуск Сервера

```bash
go run .
```

### 3. Запуск Desktop-Клиента

```bash
go run ./cmd/desktop
```

### 4. Генерация Тестового Пользователя Для Postman

```bash
env GOCACHE=/tmp/codex-gocache go run ./cmd/devkeygen -id 2
```

## Текущие Ограничения

- desktop-клиент пока хранит только один локальный профиль;
- тестирование двух пользователей на одной машине возможно, но пока не очень удобно;
- пока нет переключателя профилей;
- серверное хранилище пока остаётся in-memory и не сохраняет данные между перезапусками;
- пока нет слоя аутентификации и авторизации;
- в клиенте пока нет списка входящих и отправленных файлов.

## План Улучшения Приложения

Ниже приведён структурированный план дальнейшего развития приложения.

### Улучшения Серверной Части

- заменить in-memory репозитории на постоянное хранилище, например PostgreSQL или SQLite
- хранить зашифрованные контейнеры в реальном file storage или object storage
- добавить аутентификацию и авторизацию
- добавить проверку владельца и контроль доступа к файлам
- добавить пагинацию и endpoints для списков входящих и отправленных файлов
- добавить логирование запросов, трассировку и более подробную диагностику ошибок
- добавить конфигурацию через переменные окружения или config-файлы

### Улучшения Клиентской Части

- добавить поддержку нескольких локальных профилей
- добавить выбор активного пользователя и переключение профилей в UI
- добавить список входящих файлов и список отправленных файлов
- улучшить выбор файлов и выбор выходной директории
- отображать текущего активного локального пользователя в интерфейсе
- добавить индикаторы прогресса для длительных операций
- улучшить отображение ошибок и сценарии восстановления

### Улучшения Безопасности

- шифровать локальный профиль или хранить приватные ключи в OS keychain / secure storage
- добавить поддержку ротации ключей
- добавить отзыв ключей и сценарии замены публичных ключей
- усилить валидацию контейнеров и метаданных
- добавить версионирование формата контейнера и поддержку миграций
- усилить защиту локальных файлов профиля и прав доступа
- добавить audit logging серверных операций

### Улучшения Криптографии

- формально задокументировать протокол
- определить точные правила сериализации и гарантии совместимости
- добавить межклиентские тесты совместимости
- добавить больше негативных тестов на повреждённые и неполные контейнеры
- добавить тесты на большие файлы и проверку производительности

### Улучшения HTTP/API

- добавить OpenAPI или Swagger-документацию
- добавить интеграционные тесты полного клиент-серверного потока
- унифицировать и документировать все error responses
- ввести версионирование API
- добавить health и readiness endpoints

### Улучшения Desktop UX

- улучшить layout и визуальную иерархию интерфейса `Fyne`
- улучшить навигацию и статусные элементы
- добавить историю последних операций
- добавить drag-and-drop для отправки файлов
- более явно показывать путь сохранения расшифрованного файла

### Улучшения DevOps И Поставки

- добавить build-скрипты для сервера и desktop-клиента
- добавить CI для тестов и linting
- собирать desktop-приложение под основные платформы
- добавить release notes и versioned artifacts
- подготовить демонстрационный сценарий для защиты или презентации

### Улучшения На Уровне Продукта

- поддержка нескольких получателей для одного файла
- добавление срока действия файла или одноразовой загрузки
- хранение истории обмена
- уведомления о новых доступных файлах
- расширенное управление метаданными

## Рекомендуемые Следующие Шаги

Наиболее полезные следующие улучшения:

1. добавить поддержку нескольких профилей в desktop-клиенте
2. добавить постоянное серверное хранилище
3. добавить аутентификацию и контроль доступа
4. улучшить UX desktop-клиента вокруг отправки и получения файлов
5. подготовить проект к демонстрации и сборке в распространяемое приложение

## Текущий Статус Проекта

Проект уже находится на стадии рабочего прототипа:
- криптографический протокол реализован;
- серверный API существует и покрыт тестами;
- desktop-клиент умеет регистрировать пользователя, отправлять, получать и расшифровывать файлы;
- система ближе к корректной E2E-модели доверия, чем к server-side encryption подходу.

Следующий этап — это прежде всего развитие удобства использования, постоянного хранения, управления профилями и production-ready возможностей.

методы оптимизации/производ шифрования. 