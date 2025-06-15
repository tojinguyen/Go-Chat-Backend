ifneq (,$(wildcard ./.env))
    include .env
    export
endif

debug:
	go run ./cmd/server/main.go --debug

dev:
	go run ./cmd/server/main.go

rand-user:
	go run ./cmd/seed/main.go

lint:
	golangci-lint run

tests:
	go test -parallel=20 -covermode atomic -coverprofile=coverage.out ./...

# Unit tests only (excluding integration tests)
test-unit:
	go test -parallel=20 -covermode atomic -coverprofile=unit-coverage.out $$(go list ./... | grep -v "/tests/integration")

# Integration tests only (requires running database and Redis)
test-integration:
	go test -v -tags=integration -coverprofile=integration-coverage.out ./tests/integration/...

# Integration tests with Docker containers
test-integration-docker:
	docker-compose -f docker-compose.test.yml up -d --wait
	go test -v -tags=integration -coverprofile=integration-coverage.out ./tests/integration/... || (docker-compose -f docker-compose.test.yml down && exit 1)
	docker-compose -f docker-compose.test.yml down

# Run all tests (unit + integration)
test-all:
	$(MAKE) test-unit
	$(MAKE) test-integration

# Run migrations for testing
migrate-test:
	go run ./cmd/migrate/main.go

# Setup test database
setup-test-db:
	docker-compose -f docker-compose.test.yml up -d mysql redis --wait
	sleep 10
	$(MAKE) migrate-test

# Clean test database
clean-test-db:
	docker-compose -f docker-compose.test.yml down -v

build:
	rm ./build-out || true
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build-out cmd/main.go
	upx -9 -q ./build-out

docker build:
	docker-compose up -d --build

docker up:
	docker-compose up -d

docker down:
	docker-compose down

create-migration:
	if "$(name)" == "" ( \
		echo ❌ Thiếu tên migration. Dùng: make create-migration name=ten_migration & exit /b 1 \
	) else ( \
		goose -dir migrations/mysql create -s $(name) sql \
	)


