# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy go mod files first for better caching
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy source code
COPY src/ ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o infralog main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests (AWS, Slack, webhooks)
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 1000 infralog

# Copy binary from builder
COPY --from=builder /app/infralog /usr/local/bin/infralog

# Create directory for config and state persistence
RUN mkdir -p /etc/infralog /var/lib/infralog && \
    chown -R infralog:infralog /etc/infralog /var/lib/infralog

USER infralog

# Default config location
ENV INFRALOG_CONFIG_FILE=/etc/infralog/config.yml

ENTRYPOINT ["infralog"]
