# Build stage
FROM golang:1.21-alpine AS builder

# Install ca-certificates for HTTPS and tzdata for timezones
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations for small size
# CGO_ENABLED=0 for static binary, ldflags to strip debug info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o tg-mock \
    ./cmd/tg-mock

# Final stage - scratch for minimal image (~10MB)
FROM scratch

# Import certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /build/tg-mock /tg-mock

# Create storage directory (will be owned by root, but that's fine for scratch)
# Users should mount a volume here
VOLUME ["/data"]

# Expose the default port
EXPOSE 8081

# Run as non-root user (UID 65534 is commonly used for 'nobody')
USER 65534:65534

# Set default environment variables that map to CLI flags
ENV TG_MOCK_PORT=8081
ENV TG_MOCK_STORAGE_DIR=/data

ENTRYPOINT ["/tg-mock"]
CMD ["--port", "8081", "--storage-dir", "/data"]
