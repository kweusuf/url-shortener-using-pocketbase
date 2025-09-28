# Build stage
FROM golang:alpine AS builder

# Install necessary packages for building
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum don't change)
RUN go mod download && go mod verify

# Copy source code (cached if source files don't change)
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o url-shortener .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create app user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app/

# Copy the binary from builder stage
COPY --from=builder /app/url-shortener .

# Copy static files
COPY --chown=appuser:appgroup index.html .
COPY --chown=appuser:appgroup .env.example .env

# Copy setup script
COPY --chown=appuser:appgroup setup-admin.sh .

# Make setup script executable
RUN chmod +x setup-admin.sh

# Create directory for PocketBase data
RUN mkdir -p pb_data && chown -R appuser:appgroup pb_data

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8090

# Set environment variables
ENV POCKETBASE_DATA_DIR=/app/pb_data

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/api/system/health || exit 1

# Run the setup script for admin initialization
CMD ["./setup-admin.sh"]
