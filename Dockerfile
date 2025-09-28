# Build stage
FROM golang:latest AS builder

# Install necessary packages
RUN apt-get update && apt-get install -y \
    git \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pocketbase-demo .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create app user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app/

# Copy the binary from builder stage
COPY --from=builder /app/pocketbase-demo .

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
