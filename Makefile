.PHONY: all build clean test run docker-build docker-run help

# Variables
BINARY_NAME=cloud-resource-counter
DOCKER_IMAGE=cloud-resource-counter:latest
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Default target
all: test build

# Build the binary
build:
	@echo "Building ${BINARY_NAME}..."
	@go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/agent/main.go
	@echo "Build complete: bin/${BINARY_NAME}"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 cmd/agent/main.go
	@GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 cmd/agent/main.go
	@GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe cmd/agent/main.go
	@echo "Multi-platform build complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Tests complete. Coverage report: coverage.html"

# Run the application
run: build
	@echo "Running ${BINARY_NAME}..."
	@./bin/${BINARY_NAME} -format table

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html
	@echo "Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run
	@echo "Linting complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t ${DOCKER_IMAGE} .
	@echo "Docker build complete"

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run --rm \
		-v ${PWD}/output:/app/output \
		${DOCKER_IMAGE}

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	@mockgen -source=internal/providers/provider.go -destination=internal/mocks/provider_mock.go -package=mocks
	@echo "Mocks generated"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build        - Build the binary"
	@echo "  make build-all    - Build for multiple platforms"
	@echo "  make test         - Run tests with coverage"
	@echo "  make run          - Build and run the application"
	@echo "  make run-mock     - Run with mock data"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Lint code"
	@echo "  make deps         - Install dependencies"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run in Docker container"
	@echo "  make mocks        - Generate test mocks"
	@echo "  make help         - Show this help message"
