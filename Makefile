.PHONY: help deps unit test cover cover-docker compose-up compose-down compose-reset compose-test-up compose-test-down integration integration-testdb smoke smoke-testdb lint

help:
	@echo "Targets:"
	@echo "  deps              - go mod tidy"
	@echo "  unit              - go test ./... (unit tests)"
	@echo "  cover             - unit tests with coverage summary"
	@echo "  cover-docker      - coverage in Docker (если локальный Go сломан)"
	@echo "  compose-up        - docker compose up --build -d (без тестовой БД и без локального MinIO)"
	@echo "  compose-local-up  - docker compose --profile local-s3 up --build -d (с локальным MinIO)"
	@echo "  compose-test-up   - тестовая Postgres+migrate: docker compose --profile test up -d postgres-test migrate-test"
	@echo "  compose-test-down - остановить только сервисы профиля test"
	@echo "  compose-down      - docker compose down -v"
	@echo "  compose-reset     - down + up"
	@echo "  integration       - go test integration (БД основного compose, порт 5433)"
	@echo "  integration-testdb - то же, БД из профиля test (порт 5434)"
	@echo "  smoke             - go test smoke (порт 5433)"
	@echo "  smoke-testdb      - smoke против БД профиля test (порт 5434)"

deps:
	go mod tidy

unit:
	go test ./...

test: unit

cover:
	go test -coverprofile=cov.out ./...
	go tool cover -func=cov.out | tail -n 5

cover-docker:
	docker run --rm -v "$$PWD:/src" -w /src golang:1.25-alpine sh -lc "apk add --no-cache ca-certificates >/dev/null && go test -coverprofile=cov.out ./... && go tool cover -func=cov.out | tail -n 5"

compose-up:
	docker compose up --build -d

compose-local-up:
	docker compose --profile local-s3 up --build -d

compose-down:
	docker compose down -v

compose-reset: compose-down compose-up

compose-test-up:
	docker compose --profile test up -d postgres-test migrate-test

compose-test-down:
	docker compose --profile test stop postgres-test migrate-test

integration:
	DATABASE_URL="postgres://postgres:postgres@localhost:5433/user_service?sslmode=disable" go test -tags=integration ./integration/...

integration-testdb:
	DATABASE_URL="postgres://postgres:postgres@localhost:5434/user_service_test?sslmode=disable" go test -tags=integration ./integration/...

smoke:
	DATABASE_URL="postgres://postgres:postgres@localhost:5433/user_service?sslmode=disable" go test -tags=smoke ./smoke/...

smoke-testdb:
	DATABASE_URL="postgres://postgres:postgres@localhost:5434/user_service_test?sslmode=disable" go test -tags=smoke ./smoke/...

