.PHONY: build run test clean docker-build docker-run help

# Variables
BINARY_NAME=go-runner
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=cmd/server/main.go

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=1 go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete!"

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(BINARY_UNIX) $(MAIN_PATH)
	@echo "Linux build complete!"

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	CGO_ENABLED=1 go run $(MAIN_PATH)

# Run with hot reload (requires air)
dev:
	@echo "Running with hot reload..."
	CGO_ENABLED=1 air

# Run with custom hot reload watcher
hot-reload:
	@echo "Running with custom hot reload watcher..."
	CGO_ENABLED=1 go run cmd/hotreload/main.go

# Run with hot reload (air alternative)
dev-air:
	@echo "Running with Air hot reload..."
	CGO_ENABLED=1 air

# Test the application
test:
	@echo "Running tests..."
	go test -v ./...

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out
	@echo "Clean complete!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME) .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME)

# Docker compose up
up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Docker compose down
down:
	@echo "Stopping services..."
	docker-compose down

# Docker compose logs
logs:
	@echo "Showing logs..."
	docker-compose logs -f

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	@echo "API documentation will be available at http://localhost:8080/api/v1"

# Import example data
import-examples:
	@echo "Importing example microservices data..."
	go run scripts/import_examples.go

# Clean and setup fresh database with examples
setup-examples: clean import-examples
	@echo "Fresh setup with example data complete!"

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  build-linux    - Build for Linux"
	@echo "  run            - Run the application"
	@echo "  dev            - Run with hot reload (requires air)"
	@echo "  hot-reload     - Run with custom hot reload watcher"
	@echo "  dev-air        - Run with Air hot reload"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  up             - Start services with Docker Compose"
	@echo "  down           - Stop services"
	@echo "  logs           - Show logs"
	@echo "  docs           - Generate API documentation"
	@echo "  help           - Show this help message"
