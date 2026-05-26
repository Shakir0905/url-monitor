FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/monitor-worker \
    ./cmd/monitor-worker

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=builder /app/monitor-worker /app/monitor-worker
USER appuser
EXPOSE 9194
ENTRYPOINT ["/app/monitor-worker"]
