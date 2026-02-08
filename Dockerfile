# CERT Blockchain Node Dockerfile
# Multi-stage build for optimized image size

# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache make git gcc musl-dev linux-headers

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o /certd ./cmd/certd

# Final stage
FROM alpine:3.19

# Install runtime dependencies (including bash for entrypoint script)
RUN apk add --no-cache ca-certificates tzdata bash sed

# Create non-root user
RUN addgroup -g 1000 cert && \
    adduser -u 1000 -G cert -s /bin/sh -D cert

# Copy binary from builder
COPY --from=builder /certd /usr/local/bin/certd

# Copy entrypoint script
COPY scripts/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN sed -i 's/\r$//' /usr/local/bin/docker-entrypoint.sh && chmod +x /usr/local/bin/docker-entrypoint.sh

# Create data directory
RUN mkdir -p /root/.certd && chown -R cert:cert /root/.certd

# Set environment variables
ENV CERT_HOME=/root/.certd
ENV CHAIN_ID=cert_4283207343-1
ENV MONIKER=cert-validator

# Expose ports
# Tendermint RPC
EXPOSE 26657
# Tendermint P2P
EXPOSE 26656
# Cosmos REST API
EXPOSE 1317
# Cosmos gRPC
EXPOSE 9090
# Ethereum JSON-RPC
EXPOSE 8545
# Ethereum WebSocket
EXPOSE 8546

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

