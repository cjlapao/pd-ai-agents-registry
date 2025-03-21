FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build  -o backend cmd/server/main.go

# Use a small alpine image for the final container
FROM alpine:3.18

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/backend .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./backend"] 