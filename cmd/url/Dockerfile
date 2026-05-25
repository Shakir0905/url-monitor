FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/url-service \
    ./cmd/url

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=builder /app/url-service /app/url-service
USER appuser
EXPOSE 50052 9092
ENTRYPOINT ["/app/url-service"]
