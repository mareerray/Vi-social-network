# Backend Dockerfile
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o social-network-backend ./backend


# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache sqlite-libs ca-certificates sqlite bash

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/social-network-backend .

# Copy migrations
COPY backend/db/migrations ./db/migrations

# Copy entrypoint script if you have one
COPY backend/entrypoint.sh .
RUN chmod +x entrypoint.sh

# Create uploads directory
RUN mkdir -p uploads/avatars uploads/posts

# Expose the backend port (adjust if needed)
EXPOSE 8080

# Run the application
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./social-network-backend"]
