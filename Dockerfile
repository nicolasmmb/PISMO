# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Download modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o api cmd/api/main.go

# Runtime Stage
FROM alpine:3.18

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run
CMD ["./api"]
