# ==============================================================================
# Stage 1: Build Vue 3 Frontend (Single Page Application)
# ==============================================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend

# Install dependencies first for optimal Docker layer caching
COPY frontend/package*.json ./
RUN npm ci --prefer-offline --no-audit

# Copy frontend source and build static distribution files
COPY frontend/ ./
RUN npm run build

# ==============================================================================
# Stage 2: Compile Go Single-Binary Embedding Frontend Assets
# ==============================================================================
FROM golang:alpine AS backend-builder

WORKDIR /build

RUN apk add --no-cache ca-certificates

# Cache Go module dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY frontend/embed.go ./frontend/

# Copy built frontend assets from Stage 1 into frontend/dist for Go embed.FS
COPY --from=frontend-builder /build/frontend/dist ./frontend/dist

# Compile single binary statically (CGO-free, stripped symbols)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/lyostar ./cmd/lyostar

# ==============================================================================
# Stage 3: Minimal Production Runtime Container
# ==============================================================================
FROM alpine:3.20

# Install SSL certificates, timezone data, network utilities, and su-exec
RUN apk add --no-cache ca-certificates tzdata wget su-exec

# Create non-root user and group
RUN addgroup -g 1000 lyostar && \
    adduser -D -u 1000 -G lyostar -h /app lyostar

WORKDIR /app

# Create mount points, backward-compatible symlinks, and grant ownership
RUN mkdir -p /books /data && \
    ln -s /data /app/data && \
    ln -s /books /app/books && \
    chown -R lyostar:lyostar /books /data /app

# Copy entrypoint script and single binary
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
COPY --from=backend-builder --chown=lyostar:lyostar /app/lyostar /app/lyostar

# Storage Volumes & Port Contract
VOLUME ["/books", "/data"]
EXPOSE 8080

# Health check to ensure server is responding
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://localhost:8080/api/health || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["-books=/books", "-data=/data", "-port=8080"]
