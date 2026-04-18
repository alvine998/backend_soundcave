# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the application with cache mounts
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o backend_soundcave .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests, curl for health checks, and ffmpeg for stream processing
RUN apk --no-cache add ca-certificates curl ffmpeg

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/backend_soundcave .

# Expose port
EXPOSE 6002

# Run the application
CMD ["./backend_soundcave"]

