# Backend Dockerfile - Works with root context
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files (build context is root, so use backend/go.mod)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Strategy: Copy everything from backend, then verify cmd exists
# If cmd doesn't exist, it means backend structure is different
COPY backend/ ./

# Debug: Show what was copied
RUN echo "=== Contents of /app ===" && \
    ls -la /app/ && \
    echo "=== Checking for cmd ===" && \
    (test -d /app/cmd && echo "✓ cmd directory EXISTS" || echo "✗ cmd directory NOT FOUND") && \
    (test -d /app/backend && echo "✓ backend directory EXISTS" || echo "✗ backend directory NOT FOUND") && \
    (test -d /app/backend/cmd && echo "✓ backend/cmd EXISTS" || echo "✗ backend/cmd NOT FOUND") && \
    echo "=== All directories in /app ===" && \
    find /app -maxdepth 2 -type d | sort && \
    echo "=== Finding main.go ===" && \
    find /app -name "main.go" 2>/dev/null || echo "No main.go found"

# Build: Try to find and build main.go
RUN if [ -f /app/cmd/server/main.go ]; then \
        echo "Building from /app/cmd/server/main.go" && \
        CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server; \
    elif [ -f /app/backend/cmd/server/main.go ]; then \
        echo "Building from /app/backend/cmd/server/main.go" && \
        cd /app/backend && CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server; \
    elif [ -f /app/main.go ]; then \
        echo "Building from /app/main.go" && \
        CGO_ENABLED=0 GOOS=linux go build -o server .; \
    else \
        echo "ERROR: Cannot find main.go!" && \
        echo "Please check the file structure in AWS" && \
        find /app -type f -name "*.go" | head -10 && \
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
