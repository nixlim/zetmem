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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o amem-server ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S amem && \
    adduser -u 1001 -S amem -G amem

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/amem-server .

# Copy configuration and prompts
COPY --chown=amem:amem config/ ./config/
COPY --chown=amem:amem prompts/ ./prompts/

# Create data directory
RUN mkdir -p /app/data && chown amem:amem /app/data

# Switch to non-root user
USER amem

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD pgrep amem-server || exit 1

# Run the application
CMD ["./amem-server", "-config", "./config/docker.yaml"]
