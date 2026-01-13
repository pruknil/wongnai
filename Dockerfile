# syntax=docker/dockerfile:1

# --- Build Stage ---
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install git (required for go get)
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o server main.go

# --- Run Stage ---
FROM alpine:latest
WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/server ./server

# Set environment variables (example, override in cloud)
ENV GIN_MODE=release

# Expose port 8080 (change if your app uses a different port)
EXPOSE 8080

# Run the server
ENTRYPOINT ["./server"]

