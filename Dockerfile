# Build stage
FROM golang:1.24.3-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o server ./cmd/main.go


# =========================
# Runtime stage
# =========================
FROM alpine:3.19

# Install certs (HTTPS, Firebase, Google APIs, etc.)
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app

ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app

# Copy binary only
COPY --from=builder /app/server .

# Use non-root user
USER app

EXPOSE 3000

CMD ["./server"]
