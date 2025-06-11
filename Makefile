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

build:
	rm ./build-out || true
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build-out cmd/main.go
	upx -9 -q ./build-out

docker build:
	docker-compose up -d --build

docker up:
	docker-compose up -d

create-migration:
	if "$(name)" == "" ( \
		echo ❌ Thiếu tên migration. Dùng: make create-migration name=ten_migration & exit /b 1 \
	) else ( \
		goose -dir migrations/mysql create -s $(name) sql \
	)


