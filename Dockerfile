# Build stage
FROM golang:1.24-alpine AS builder

# Install git (if needed for go mod) and other dependencies
RUN apk update && apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum to download dependencies early caching them
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code.
COPY . .

# Build the application statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cobalt-discord-app .

# Final stage
FROM alpine:latest

# Copy the built binary from the builder stage.
COPY --from=builder /app/cobalt-discord-app /cobalt-discord-app

# Use a non-root user if needed, for now executing as root in scratch
ENTRYPOINT ["/cobalt-discord-app"]
