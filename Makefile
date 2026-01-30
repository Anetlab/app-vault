.PHONY: help build run test clean docker-up docker-down migrate

help:
	@echo "SecureVault - Available commands:"
	@echo "  make build       - Build the server binary"
	@echo "  make run         - Run the server"
	@echo "  make test        - Run all tests"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make docker-up   - Start PostgreSQL with docker-compose"
	@echo "  make docker-down - Stop PostgreSQL"
	@echo "  make migrate     - Run database migrations"

build:
	go build -o bin/securevault cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test ./... -v

clean:
	rm -rf bin/

docker-up:
	docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

docker-down:
	docker-compose down

migrate: docker-up
	@echo "Running migrations..."
	@go run cmd/server/main.go &
	@sleep 2
	@pkill -f "cmd/server/main.go"
	@echo "Migrations completed"
