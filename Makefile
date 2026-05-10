.PHONY: build run test e2e lint  clean

# Сборка
build:
	go build -o bin/dungeon cmd/app/main.go

# Запуск с дефолтными файлами
run:
	go run cmd/app/main.go -config config.json -events events

# Запуск юнит тестов 
test:
	go test -v ./internal/...

# Запуск E2E тестов
e2e:
	go test -v ./tests/...

# golangci-lint
lint:
	golangci-lint run

# Чистка скомпилированных файлов
clean:
	rm -rf bin/