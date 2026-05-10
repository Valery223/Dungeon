.PHONY: build run test e2e lint  clean

CONFIG ?= config.json
EVENTS ?= events

all: lint unit-test build

# Сборка
build:
	@echo "==> Building application..."
	go build -o bin/dungeon cmd/app/main.go

# Запуск с дефолтными файлами
run:
	@echo "==> Running application..."
	go run cmd/app/main.go -config config.json -events events

test: unit-test

# Запуск юнит тестов 
unit-test:
	@echo "==> Running unit tests..."
	go test -race -v -short ./internal/...

load-test:
	@echo "==> Running load tests..."
	go test -v -run TestGameRunner_Load1MillionEvents ./internal/usecase/...

# Запуск E2E тестов
e2e:
	@echo "==> Running E2E tests..."
	go test -race -v ./tests/...

# golangci-lint
lint:
	golangci-lint run

docker-build:
	@echo "==> Building Docker image..."
	docker build -t dungeon-app .

docker-run: docker-build
	@echo "==> Running in Docker..."
	docker run --rm \
		-v $(shell pwd)/$(CONFIG):/app/config.json \
		-v $(shell pwd)/$(EVENTS):/app/events\
		dungeon-app

docker-run-force:
	docker run --rm \
		-v $(shell pwd)/$(CONFIG):/app/config.json \
		-v $(shell pwd)/$(EVENTS):/app/events \
		dungeon-app

# Чистка скомпилированных файлов
clean:
	rm -rf bin/