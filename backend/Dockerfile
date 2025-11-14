# Backend Dockerfile - Works with root context
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files (build context is root, so use backend/go.mod)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code - explicitly copy cmd directory first
# cmd directory might be ignored, so copy it explicitly
COPY backend/cmd ./cmd
COPY backend/internal ./internal

# Debug: Verify what was copied
RUN echo "=== Root directory contents ===" && \
    ls -la /app/ && \
    echo "=== cmd directory ===" && \
    ls -la /app/cmd/ && \
    echo "=== cmd/server directory ===" && \
    ls -la /app/cmd/server/ && \
    echo "=== main.go file ===" && \
    ls -la /app/cmd/server/main.go

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
