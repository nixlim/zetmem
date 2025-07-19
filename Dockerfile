# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o zetmem-server ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S zetmem && \
    adduser -u 1001 -S zetmem -G zetmem

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/zetmem-server .

# Copy configuration and prompts
COPY --chown=zetmem:zetmem config/ ./config/
COPY --chown=zetmem:zetmem prompts/ ./prompts/

# Create data directory
RUN mkdir -p /app/data && chown zetmem:zetmem /app/data

# Switch to non-root user
USER zetmem

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD pgrep zetmem-server || exit 1

# Run the application
CMD ["./zetmem-server", "-config", "./config/docker.yaml"]
