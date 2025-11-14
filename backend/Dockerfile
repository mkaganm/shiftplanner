# Backend Dockerfile - Works with root context
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files (build context is root, so use backend/go.mod)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code - explicitly copy each directory
COPY backend/cmd ./cmd
COPY backend/internal ./internal

# Debug: Verify directory structure BEFORE build
RUN echo "=== Root directory ===" && ls -la /app/ && \
    echo "=== cmd directory ===" && ls -la /app/cmd/ && \
    echo "=== cmd/server directory ===" && ls -la /app/cmd/server/ && \
    echo "=== cmd/server/main.go ===" && ls -la /app/cmd/server/main.go && \
    echo "=== Finding all .go files in cmd ===" && find /app/cmd -name "*.go" 2>/dev/null || echo "No .go files found"

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV DATA_DIR=/data

# Run the server
CMD ["./server"]
