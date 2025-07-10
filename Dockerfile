FROM golang:1.23-alpine AS builder

# Install git (required for git operations)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rebAIser cmd/rebAIser/main.go

# Final stage
FROM alpine:3.19

# Install git and ca-certificates (required for git operations and HTTPS)
RUN apk add --no-cache git ca-certificates openssh-client

# Create a non-root user
RUN adduser -D -s /bin/sh rebuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/rebAIser .

# Make binary executable
RUN chmod +x rebAIser

# Change ownership to non-root user
RUN chown rebuser:rebuser /app/rebAIser

# Switch to non-root user
USER rebuser

# Set entrypoint
ENTRYPOINT ["./rebAIser"]