# Backend Dockerfile - Optimized with scratch (ultra-lightweight, ~5MB vs Alpine ~5MB)
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

# Final stage - scratch (empty image, only binary, ~0MB base + binary size)
FROM scratch

# Copy ca-certificates from alpine for HTTPS support
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy only the binary from builder
COPY --from=builder /app/server /server

# Set environment variables
ENV PORT=8080 \
    DATA_DIR=/data

# Expose port
EXPOSE 8080

# Run the server
ENTRYPOINT ["/server"]
