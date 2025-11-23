# Backend Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

COPY --from=builder /app/server .

ENV PORT=8080 DATA_DIR=/data

EXPOSE 8080

RUN mkdir -p /data && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /root

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:${PORT}/api/holidays || exit 1

# Fix permissions and switch to appuser
ENTRYPOINT ["sh", "-c", "chown -R appuser:appuser /data && chmod -R 755 /data && su appuser -c './server'"]
