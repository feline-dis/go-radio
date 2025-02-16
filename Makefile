# Variables
BINARY_NAME=go.radio
VERSION?=1.0.0
BUILD_DIR=build
GO_BUILD_FLAGS=-ldflags="-w -s -X main.Version=${VERSION}"

# Docker compose files
COMPOSE_FILES=-f docker-compose.yml

.PHONY: all build clean test docker-build docker-push run help

# Default target
all: clean build

## Build:
build: build-backend build-frontend ## Build both backend and frontend

build-backend: ## Build the backend binary
	@echo "Building backend..."
	@ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build ${GO_BUILD_FLAGS} -o ./${BUILD_DIR}/${BINARY_NAME}

build-frontend: ## Build the frontend
	@echo "Building frontend..."
	@cd public_react && yarn && yarn build 

## Testing:
test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	@echo "Testing backend..."
	@cd backend && go test -v ./...

test-frontend: ## Run frontend tests
	@echo "Testing frontend..."
	@cd frontend && npm test

## Docker:
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker compose ${COMPOSE_FILES} build

docker-push: ## Push Docker images to registry
	@echo "Pushing Docker images..."
	docker compose ${COMPOSE_FILES} push

## Development:
dev: ## Start development environment
	docker compose ${COMPOSE_FILES} up --build

dev-down: ## Stop development environment
	docker compose ${COMPOSE_FILES} down

## Production:
prod: ## Start production environment
	docker compose ${COMPOSE_FILES} -f docker-compose.prod.yml up -d

prod-down: ## Stop production environment
	docker compose ${COMPOSE_FILES} -f docker-compose.prod.yml down

## Clean:
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf ${BUILD_DIR}
	@cd frontend && rm -rf build node_modules
	@cd backend && rm -rf vendor

## Utility:
lint: ## Run linters
	@echo "Linting backend..."
	@cd backend && golangci-lint run
	@echo "Linting frontend..."
	@cd frontend && npm run lint

## Dependencies:
deps: ## Install development dependencies
	@echo "Installing backend dependencies..."
	@cd backend && go mod download
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install

## Help:
help: ## Show this help
	@echo "Makefile targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Create necessary directories
$(shell mkdir -p ${BUILD_DIR})
