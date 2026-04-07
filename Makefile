.PHONY: dev dev-backend dev-frontend build build-frontend build-backend clean test lint docker

GOBIN := $(shell go env GOPATH)/bin

# Development: run backend and frontend with hot-reload
dev:
	@echo "Starting development servers..."
	@make -j2 dev-backend dev-frontend

dev-backend:
	cd web && npm run build
	$(GOBIN)/air -c .air.toml

dev-frontend:
	cd web && npm run dev

# Build production binary
build: build-frontend build-backend

build-frontend:
	cd web && npm install && npm run build

build-backend:
	go build -ldflags="-s -w" -o bin/nats-insight ./cmd/nats-insight/

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf cmd/nats-insight/web/assets/
	rm -f cmd/nats-insight/web/index.html
	rm -f cmd/nats-insight/web/favicon.svg
	rm -rf web/node_modules/
	rm -rf tmp/

# Run tests
test:
	go test ./...
	cd web && npm run test 2>/dev/null || true

# Run linters
lint:
	golangci-lint run ./...
	cd web && npm run lint 2>/dev/null || true

# Docker
docker:
	docker compose -f docker/docker-compose.yml build

docker-up:
	docker compose -f docker/docker-compose.yml up

docker-down:
	docker compose -f docker/docker-compose.yml down
