# Build stage
FROM golang:1.24-alpine AS builder

# Install Node.js and npm for node_modules
RUN apk add --no-cache nodejs npm

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download go dependencies
RUN go mod download

# Copy package.json and install node dependencies
COPY package.json ./
RUN npm install

# Copy source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/web

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy static files and templates
COPY --from=builder /app/ui ./ui
COPY --from=builder /app/node_modules ./node_modules

# Expose port (will be set by DigitalOcean)
EXPOSE 8080

# Command to run the application
CMD ["./main"]
