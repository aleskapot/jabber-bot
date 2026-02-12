# Multi-stage Dockerfile for Jabber Bot

# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=1.0.0
ARG BUILD_TIME
ARG GIT_COMMIT

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
#    -ldflags="-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.GitCommit=${GIT_COMMIT}'" \
    -a -installsuffix cgo \
    -o bin/jabber-bot \
    ./cmd/server

# Production stage
FROM alpine:latest

# Environment variables
ENV JABBER_BOT_LOG_LEVEL=info
ENV JABBER_BOT_API_HOST=0.0.0.0
ENV JABBER_BOT_API_PORT=8080
ENV TZ=Europe/Moscow

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    echo "${TZ}" && \
    date

# Create non-root user
RUN addgroup -g 1001 -S jabber && \
    adduser -u 1001 -S jabber -G jabber

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/jabber-bot .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Create directories for logs and data
RUN mkdir -p /app/logs /app/data && \
    chown -R jabber:jabber /app

# Switch to non-root user
USER jabber

# Expose ports
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Default command
CMD ["./jabber-bot", "-config", "configs/config.yaml"]