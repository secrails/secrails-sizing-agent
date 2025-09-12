# Build stage   
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty) \
    -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') \
    -X main.GitCommit=$(git rev-parse --short HEAD)" \
    -o cloud-resource-counter \
    cmd/agent/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cloud-resource-counter /app/
COPY --from=builder /build/configs /app/configs

# Create output directory
RUN mkdir -p /app/output && chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Set entrypoint
ENTRYPOINT ["/app/cloud-resource-counter"]

# Default command
CMD ["-config", "/app/configs/config.yaml", "-format", "json", "-output", "/app/output/report.json"]