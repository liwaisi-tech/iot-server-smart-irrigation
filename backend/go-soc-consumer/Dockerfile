# Build stage
FROM golang:1.23.11-alpine AS builder

# Install build dependencies and security tools
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    upx \
    wget \
    && update-ca-certificates

# Create build cache directory
RUN mkdir -p /go/pkg/mod

# Set working directory
WORKDIR /app

# Copy go mod files for dependency caching
COPY go.mod go.sum ./

# Download dependencies with module cache
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && \
    go mod verify

# Copy source code (exclude unnecessary files)
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Build optimized binary with security flags
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -trimpath \
    -o server ./cmd/server

# Compress binary to reduce size
RUN upx --best --lzma server

# Security scanner stage (optional, uncomment for production)
# FROM aquasec/trivy:latest AS security
# COPY --from=builder /app/server /app/server
# RUN trivy filesystem --exit-code 1 --no-progress /app/server

# Final stage - use distroless for maximum security
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data, CA certificates, and wget for health checks
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/bin/wget /usr/bin/wget

# Set working directory
WORKDIR /app

# Copy the optimized binary
COPY --from=builder /app/server ./server

# Use distroless nonroot user (uid:gid 65532:65532)
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Health check using wget to ping endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD ["/usr/bin/wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/ping"]

# Set resource limits via environment
ENV GOGC=100 \
    GOMEMLIMIT=512MiB \
    GOMAXPROCS=2

# Command to run the application
CMD ["./server"]