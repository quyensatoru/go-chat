# Build stage
FROM golang:1.24.3-alpine AS builder

# Install git for fetching dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o server ./cmd/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy the Pre-built binary from the previous stage
COPY --from=builder /app/server .
COPY --from=builder /app/.env . 

# Expose port 3000 to the outside world
EXPOSE 3000

# Command to run the executable
CMD ["./server"]
