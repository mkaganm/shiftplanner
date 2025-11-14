# Backend Dockerfile - Works with root context
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files (build context is root, so use backend/go.mod)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./

# Build the application
RUN if [ -f /app/cmd/server/main.go ]; then \
        CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server; \
    elif [ -f /app/backend/cmd/server/main.go ]; then \
        cd /app/backend && CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server; \
    elif [ -f /app/main.go ]; then \
        CGO_ENABLED=0 GOOS=linux go build -o server .; \
    else \
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
