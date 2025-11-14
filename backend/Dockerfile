# Backend Dockerfile - Works with root context
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Debug: List what's available in build context
RUN echo "=== Build context check ===" && ls -la / || echo "Cannot list root"

# Copy go mod files (build context is root, so use backend/go.mod)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code - try different approaches
# First, copy everything from backend
COPY backend/ ./

# Debug: Verify what was copied
RUN echo "=== After COPY backend/ ===" && \
    ls -la /app/ && \
    echo "=== cmd directory ===" && \
    (ls -la /app/cmd/ || echo "cmd not found") && \
    echo "=== Finding cmd ===" && \
    find /app -type d -name "cmd" 2>/dev/null || echo "No cmd found"

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
