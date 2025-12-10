# syntax=docker/dockerfile:1

# Build stage (match required Go version)
FROM golang:1.25.4-alpine AS builder
WORKDIR /app

# Tools for static build + compression
RUN apk add --no-cache ca-certificates git upx

# Cache go env and dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary with optimizations and compress with upx
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -extldflags '-static'" -a -o /app/sso-api ./cmd/api \
    && upx --best --lzma /app/sso-api

# Runtime stage (scratch for minimal image)
FROM scratch
WORKDIR /app

# Certs for HTTPS calls (if any)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary only (statically linked & compressed)
COPY --from=builder /app/sso-api /app/sso-api

# Default values can be overridden at runtime
ENV PORT=:8080

EXPOSE 8080
USER 65532

CMD ["/app/sso-api"]
