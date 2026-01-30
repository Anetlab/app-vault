# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Ensure migrations directory exists (create empty if not present)
RUN mkdir -p migrations

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o appvault ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates postgresql-client

# Create non-root user
RUN addgroup -g 1000 appvault && \
    adduser -D -u 1000 -G appvault appvault

WORKDIR /home/appvault

# Copy binary from builder
COPY --from=builder /app/appvault .

# Copy migrations directory from builder (will exist even if empty)
COPY --from=builder --chown=appvault:appvault /app/migrations ./migrations

# Create certs directory
RUN mkdir -p ./certs && chown -R appvault:appvault /home/appvault

# Switch to non-root user
USER appvault

# Expose port
EXPOSE 8080 8443

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./appvault"]
