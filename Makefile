.PHONY: build run test e2e lint  clean

CONFIG ?= config.json
EVENTS ?= events

# Сборка
build:
	go build -o bin/dungeon cmd/app/main.go

# Запуск с дефолтными файлами
run:
	go run cmd/app/main.go -config config.json -events events

# Запуск юнит тестов 
test:
	go test -race -v ./internal/...

# Запуск E2E тестов
e2e:
	go test -race -v ./tests/...

# golangci-lint
lint:
	golangci-lint run

docker-build:
	@echo "==> Building Docker image..."
	docker build -t dungeon-app .

docker-run: docker-build
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