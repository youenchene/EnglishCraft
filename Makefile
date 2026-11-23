.PHONY: build dev test test-unit test-integration test-architecture test-frontend test-e2e lint audit docker-up docker-down

db:
	docker compose up -d --wait postgres

migrate:
	APP_COMMAND=migrate go run ./cmd/englishcraft

provision:
	APP_COMMAND=provision go run ./cmd/englishcraft

api:
	go run ./cmd/englishcraft

web:
	cd web && npm run dev -- --host 0.0.0.0

build:
	cd web && npm ci && npm run build
	go build -o bin/englishcraft ./cmd/englishcraft

dev:
	docker compose up --build

test: test-unit test-architecture test-frontend

test-unit:
	go test ./internal/...

test-integration:
	TEST_DATABASE_URL=$${TEST_DATABASE_URL:-postgres://englishcraft:local-development-only@localhost:5432/englishcraft?sslmode=disable} go test -tags=integration ./test/integration/...

test-architecture:
	go test ./test/architecture/...

test-frontend:
	cd web && npm test

test-e2e:
	cd web && npm run test:e2e

lint:
	test -z "$$(gofmt -l internal cmd content test)"
	go vet ./...
	go test ./test/architecture/...
	cd web && npm run lint

audit:
	cd web && npm audit

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down
