# Backend Dockerfile - Works with root context
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files (build context is root, so use backend/go.mod)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy entire backend directory (this should include cmd/)
COPY backend/ ./

# Debug: Show what was actually copied
RUN echo "=== Contents of /app ===" && \
    ls -la /app/ && \
    echo "=== Looking for cmd ===" && \
    find /app -maxdepth 2 -type d -name "cmd" 2>/dev/null && \
    echo "=== Looking for main.go ===" && \
    find /app -name "main.go" 2>/dev/null && \
    echo "=== Full directory structure ===" && \
    find /app -type d -maxdepth 3 2>/dev/null | sort

# Build - find main.go and build from there
RUN MAIN_GO=$(find /app -name "main.go" -path "*/cmd/server/*" 2>/dev/null | head -1) && \
    if [ -n "$MAIN_GO" ]; then \
        MAIN_DIR=$(dirname "$MAIN_GO") && \
        echo "Found main.go at: $MAIN_GO" && \
        echo "Building from directory: $MAIN_DIR" && \
        cd "$MAIN_DIR" && \
        CGO_ENABLED=0 GOOS=linux go build -o /app/server .; \
    else \
        echo "ERROR: main.go not found in cmd/server!" && \
        echo "All .go files found:" && \
        find /app -name "*.go" 2>/dev/null | head -10 && \
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
