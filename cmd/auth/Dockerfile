# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source.
COPY . .

# Build the binary.
# -ldflags reduces binary size, CGO_ENABLED=0 makes it fully static.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/auth-service \
    ./cmd/auth

# Stage 2: Runtime
FROM alpine:3.20

# Add CA certificates for HTTPS calls and tzdata for timezones.
RUN apk --no-cache add ca-certificates tzdata

# Run as non-root user for security.
RUN adduser -D -u 1000 appuser

WORKDIR /app

COPY --from=builder /app/auth-service /app/auth-service

USER appuser

EXPOSE 50051 9091

ENTRYPOINT ["/app/auth-service"]

