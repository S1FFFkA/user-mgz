# user-mgz

Микросервис пользователей на Go + gRPC.

Технические решения:
- PostgreSQL драйвер: `pgx` (`pgxpool`)
- ID пользователей: UUIDv7 (генерируются в приложении)
- Фото хранятся в S3, в БД сохраняются метаданные (`object_key` / `url`)
- Главная фото хранится в `users`, дополнительные фото в `user_photos` (позиции 1..6)

## Что уже заложено

- Контракт gRPC в `gRPC/service.proto`
- gRPC сервер в `cmd/main.go`
- Реализованные gRPC handlers в `internal/transport/grpc/user/server.go`
- Слоистый каркас:
  - `internal/storage/postgres` (инициализация пула PostgreSQL)
  - `internal/storage/s3` (инициализация S3 клиента)
  - `internal/repository/user`, `internal/repository/userphoto`, `internal/repository/s3` (SQL/S3-реализация по доменам)
  - `internal/repository/interfaces.go` (интерфейсы)
  - `internal/usecase/user` (бизнес-слой/use-case через интерфейсы репозиториев)
  - `internal/domain` (домен-модели)
  - `pkg/logger` (`zap` JSON логгер)

## Методы сервиса (MVP)

- `CreateUser` - создать пользователя
- `GetUser` - получить пользователя по `user_id`
- `UpdateUser` - обновить пользователя
- `DeleteUser` - удалить пользователя
- `ListUsers` - список пользователей (`limit`, `offset`, фильтр по `city_id`)
- `GetUserPhotoUploadUrl` - получить presigned URL для загрузки фото в S3
- `ConfirmUserPhotoUpload` - подтвердить загрузку фото в S3 и обновить БД
- `DeleteUserPhoto` - удаление extra-фото запрещено бизнес-правилом (возвращается ошибка)
- `GetUserPhotoDownloadUrl` - получить presigned URL для скачивания фото

## Установка зависимостей для генерации

```powershell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Убедись, что `%USERPROFILE%\go\bin` есть в `PATH`.

## Генерация protobuf/gRPC кода

Запускать из корня проекта:

```powershell
protoc -I . -I C:\protoc\include --go_out=. --go_opt=module=github.com/S1FFFkA/user-mgz --go-grpc_out=. --go-grpc_opt=module=github.com/S1FFFkA/user-mgz gRPC/service.proto
```

После этого сгенерируются файлы:

- `pkg/api/user/v1/service.pb.go`
- `pkg/api/user/v1/service_grpc.pb.go`

## SQL схема

Схема хранится в миграциях `migrations/`:

- `001_init.up.sql` — применить всю схему
- `001_init.down.sql` — откатить схему

- `cities`
- `users`
- `user_photos`
- enum `sex` (`male`, `female`)
- триггер на `users.updated_at`

Применение миграции вручную:

```powershell
psql "$env:DATABASE_URL" -f .\migrations\001_init.up.sql
```

Откат:

```powershell
psql "$env:DATABASE_URL" -f .\migrations\001_init.down.sql
```

Важно: в схеме `users.id` без `DEFAULT`, потому что UUIDv7 генерируется в Go-коде репозитория.

## Локальный запуск без Docker

```powershell
go mod tidy
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/user_service?sslmode=disable"
$env:S3_ENDPOINT="s3.buckets.ru"
$env:S3_ACCESS_KEY="<access_key>"
$env:S3_SECRET_KEY="<secret_key>"
$env:S3_BUCKET="fotos"
$env:S3_USE_SSL="true"
go run .\cmd
```

По умолчанию сервер слушает `:50051`. Порт можно задать через `GRPC_PORT`.
Обязательные env: `DATABASE_URL`, `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET`.

Минимум репозиторных интерфейсов разделен на:

- `UserRepository` — CRUD/List пользователей
- `UserPhotoRepository` — операции с фото-метаданными в Postgres
- `S3Repository` — presigned URL и удаление объектов в S3

## Запуск через Docker Compose

Поднимаются 3 контейнера:
- `postgres` (БД)
- `migrate` (одноразовый контейнер, применяет `migrations/*.up.sql` и завершается)
- `app` (gRPC сервис)

Скопируй пример env и заполни ключи:

```powershell
Copy-Item .\env.example .\.env
```

Запуск:

```powershell
docker compose up --build
```

Остановить:

```powershell
docker compose down
```

Остановить и удалить volume с данными БД:

```powershell
docker compose down -v
```

## Smoke-тест клиента

Для проверки есть минимальный gRPC клиент `cmd/client`.

Пример сценария:

```powershell
# 1) Создать пользователя
go run ./cmd/client create-user --first-name Ivan --last-name Petrov --email ivan@example.com --birth-date 2000-01-01 --toiler 7 --sex male --primary-key users/bootstrap/primary.jpg --primary-url users/bootstrap/primary.jpg

# 2) Получить upload URL для extra фото (позиция 1)
go run ./cmd/client photo-upload-url --user-id <user_id> --type extra --extra-pos 1 --content-type image/jpeg --content-length 123456

# 3) Подтвердить загрузку фото
go run ./cmd/client photo-confirm --user-id <user_id> --type extra --extra-pos 1 --object-key <object_key>

# 4) Получить пользователя
go run ./cmd/client get-user --user-id <user_id>
```

Порт клиента задается через `USER_GRPC_ADDR`, по умолчанию `localhost:50051`.

## Тесты

Unit:

```powershell
go test ./internal/usecase/user ./internal/transport/grpc/user -cover
```

Integration (требует Docker):

```powershell
go test -tags=integration ./internal/repository/user ./internal/repository/userphoto -cover
```


