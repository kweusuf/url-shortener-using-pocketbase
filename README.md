# URL Shortener - Production Deployment Guide

A robust URL shortener application built with Go, PocketBase, and modern web technologies.

## 🚀 Production Deployment Strategy

### Environment Configuration

The application supports multiple deployment strategies through environment variables:

#### 1. Environment Variables

```bash
# Required: Set your production domain
export BASE_URL="https://yourdomain.com"

# Optional: Database configuration
export DATABASE_URL="your-database-connection-string"

# Optional: Port configuration
export PORT=8080
```

#### 2. Supported Protocols

- **HTTP/HTTPS**: Automatically detected and configured
- **WebSocket/WSS**: Automatically converted based on HTTP protocol
- **Custom Domains**: Full support for any domain name

### Deployment Options

#### Option 1: Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o url-shortener .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/url-shortener .
COPY --from=builder /app/index.html .
COPY --from=builder /app/pb_data ./pb_data

# Set production environment
ENV BASE_URL="https://yourdomain.com"
ENV PORT=8080

EXPOSE 8080
CMD ["./url-shortener", "serve"]
```

#### Option 2: Systemd Service

```ini
[Unit]
Description=URL Shortener Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/url-shortener
Environment=BASE_URL=https://yourdomain.com
Environment=PORT=8080
ExecStart=/opt/url-shortener/url-shortener serve
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

#### Option 3: Reverse Proxy (Recommended)

**Nginx Configuration:**
```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8090;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # WebSocket support
    location /ws {
        proxy_pass http://localhost:8090;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### URL Configuration Functions

The application includes several helper functions for dynamic URL configuration:

```go
// Get base URL from environment or default
func getBaseURL() string {
    if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
        return strings.TrimSuffix(baseURL, "/")
    }
    return "http://localhost:8090"
}

// Get API base URL
func getAPIBaseURL() string {
    return getBaseURL() + "/api"
}

// Get WebSocket URL with protocol conversion
func getWebSocketURL() string {
    baseURL := getBaseURL()
    if strings.HasPrefix(baseURL, "https://") {
        return "wss://" + strings.TrimPrefix(baseURL, "https://") + "/ws"
    }
    return "ws://" + strings.TrimPrefix(baseURL, "http://") + "/ws"
}
```

### Frontend Configuration

The frontend automatically detects the current domain:

```javascript
// Dynamic URL configuration
const BASE_URL = window.location.origin;
const API_BASE = BASE_URL + '/api';
const WS_URL = (BASE_URL.startsWith('https:') ? 'wss:' : 'ws:') +
               '//' + window.location.host + '/ws';
```

## 🔧 Production Checklist

### Pre-Deployment
- [ ] Set `BASE_URL` environment variable
- [ ] Configure SSL/TLS certificates
- [ ] Set up reverse proxy (nginx/recommended)
- [ ] Configure firewall rules
- [ ] Set up monitoring and logging
- [ ] Configure backup strategy

### Security Considerations
- [ ] Enable HTTPS in production
- [ ] Set secure headers
- [ ] Configure CORS properly
- [ ] Rate limiting for API endpoints
- [ ] Input validation and sanitization
- [ ] Database security best practices

### Performance Optimization
- [ ] Enable database connection pooling
- [ ] Configure caching headers
- [ ] Optimize static file serving
- [ ] Set up CDN for static assets
- [ ] Database indexing strategy

### Monitoring & Maintenance
- [ ] Set up health check endpoints
- [ ] Configure log aggregation
- [ ] Set up alerting for errors
- [ ] Regular backup schedules
- [ ] Database cleanup automation

## 🌐 Domain Configuration Examples

### Single Domain
```bash
export BASE_URL="https://short.ly"
```

### Subdomain
```bash
export BASE_URL="https://url.yourcompany.com"
```

### Custom Port
```bash
export BASE_URL="https://yourdomain.com:8080"
```

### Load Balancer
```bash
export BASE_URL="https://urls.company.com"
```

## 🔄 CI/CD Pipeline Example

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production
on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Build and Deploy
      run: |
        docker build -t url-shortener .
        docker run -d \
          --name url-shortener \
          -e BASE_URL="https://yourdomain.com" \
          -p 8080:8080 \
          url-shortener
```

## 📊 Health Check Endpoints

- **Application Health**: `GET /api/hello`
- **Database Status**: `GET /api/recent`
- **WebSocket Test**: `WS /ws`

## 🚀 Quick Start Commands

### Development
```bash
go run main.go serve
# Access at http://localhost:8090
```

### Production (Docker)
```bash
docker run -d \
  -e BASE_URL="https://yourdomain.com" \
  -p 8080:8080 \
  your-image
```

### Production (Systemd)
```bash
sudo systemctl enable url-shortener
sudo systemctl start url-shortener
```

## 🔍 Troubleshooting

### Common Issues

1. **URLs showing localhost in production**
   - Check `BASE_URL` environment variable
   - Verify reverse proxy configuration

2. **WebSocket connection failures**
   - Ensure proper proxy configuration for `/ws`
   - Check SSL/TLS certificate validity

3. **CORS issues**
   - Configure proper CORS headers
   - Check origin validation

### Debug Commands

```bash
# Check environment variables
echo $BASE_URL

# Test API endpoints
curl https://yourdomain.com/api/hello

# Check application logs
journalctl -u url-shortener -f
```

## 📝 License

This project is open source and available under the MIT License.
