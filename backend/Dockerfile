# Backend Dockerfile - Works with root context
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files (build context is root, so use backend/go.mod)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# IMPORTANT: Copy cmd directory FIRST and explicitly
# This ensures cmd/ is not excluded by .dockerignore
COPY backend/cmd ./cmd

# Copy rest of backend
COPY backend/internal ./internal

# Debug: Show what was actually copied
RUN echo "=== Contents of /app ===" && \
    ls -la /app/ && \
    echo "=== cmd directory ===" && \
    ls -la /app/cmd/ 2>/dev/null || echo "cmd NOT FOUND" && \
    echo "=== cmd/server directory ===" && \
    ls -la /app/cmd/server/ 2>/dev/null || echo "cmd/server NOT FOUND" && \
    echo "=== main.go file ===" && \
    ls -la /app/cmd/server/main.go 2>/dev/null || echo "main.go NOT FOUND"

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
