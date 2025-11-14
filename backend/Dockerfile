# Backend Dockerfile - Works with root context
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files (build context is root, so use backend/go.mod)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code - copy entire backend directory
COPY backend/ ./

# Debug: Verify what was copied (this will show in build output)
RUN echo "=== Root directory contents ===" && \
    ls -la /app/ && \
    echo "=== Checking for cmd ===" && \
    (test -d /app/cmd && echo "cmd directory EXISTS" || echo "cmd directory NOT FOUND") && \
    (test -d /app/backend/cmd && echo "backend/cmd directory EXISTS" || echo "backend/cmd directory NOT FOUND") && \
    echo "=== Finding all cmd directories ===" && \
    find /app -type d -name "cmd" 2>/dev/null && \
    echo "=== Finding all main.go files ===" && \
    find /app -type f -name "main.go" 2>/dev/null && \
    echo "=== Listing all directories in /app ===" && \
    find /app -type d -maxdepth 3 2>/dev/null | head -20

# Determine the correct path for cmd/server
# Try multiple possible locations
RUN if [ -d /app/cmd/server ]; then \
        echo "Building from /app/cmd/server" && \
        CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server; \
    elif [ -d /app/backend/cmd/server ]; then \
        echo "Building from /app/backend/cmd/server" && \
        cd /app/backend && CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server; \
    elif [ -f /app/cmd/server/main.go ]; then \
        echo "Building from /app/cmd/server/main.go" && \
        CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server; \
    elif [ -f /app/backend/cmd/server/main.go ]; then \
        echo "Building from /app/backend/cmd/server/main.go" && \
        cd /app/backend && CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server; \
    else \
        echo "ERROR: cmd/server not found!" && \
        echo "=== Full directory tree ===" && \
        find /app -type d 2>/dev/null | head -30 && \
        echo "=== All .go files ===" && \
        find /app -type f -name "*.go" 2>/dev/null | head -20 && \
        exit 1; \
    fi

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
