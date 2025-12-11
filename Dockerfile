# syntax=docker/dockerfile:1.7-labs

# Build stage (Go version as declared in go.mod)
FROM golang:1.25.4-alpine AS builder
WORKDIR /app

# Minimal tooling; avoid extra packages to keep build light on small VMs
RUN apk add --no-cache ca-certificates

ENV CGO_ENABLED=0 GOOS=linux

# Cache modules and build artifacts between builds to reduce CPU/RAM thrash
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go env -w GOMODCACHE=/go/pkg/mod

# Resolve deps first for better layer reuse
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source
COPY . .

# Build optimized static binary; drop debug info and VCS metadata
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /app/sso-api ./cmd/api

# Runtime stage (scratch for minimal image)
FROM scratch
WORKDIR /app

# Certs for HTTPS calls (if any)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary only (statically linked)
COPY --from=builder /app/sso-api /app/sso-api

# Default values can be overridden at runtime
ENV PORT=:8080

EXPOSE 8080
USER 65532

CMD ["/app/sso-api"]
