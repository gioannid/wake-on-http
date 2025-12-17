# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download || true

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wake-on-http .

# Final stage - lightweight runtime
FROM alpine:latest

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /build/wake-on-http .

# Run the application
CMD ["./wake-on-http"]
