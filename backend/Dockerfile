# Backend Dockerfile - Optimized with alpine (minimal + healthcheck support)
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching
COPY backend/go.mod backend/go.sum ./

# Download dependencies with cache mount for faster rebuilds
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy backend source code
COPY backend/ ./

# Build static binary with all optimizations
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -trimpath \
    -tags netgo \
    -installsuffix netgo \
    -o server ./cmd/server

# Final stage - alpine minimal (supports healthcheck)
FROM alpine:latest

# Install only ca-certificates and wget for healthcheck
RUN apk --no-cache add ca-certificates wget && \
    rm -rf /var/cache/apk/*

WORKDIR /root/

# Copy only the binary from builder
COPY --from=builder /app/server .

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080 \
    DATA_DIR=/data

# Health check - use /api/holidays endpoint (unprotected)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:${PORT}/api/holidays || exit 1

# Use non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /root /data

USER appuser

# Run the server
ENTRYPOINT ["./server"]
