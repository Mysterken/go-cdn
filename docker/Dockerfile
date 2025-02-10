# Use the official Go image as a build environment
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and install dependencies
COPY ../go.mod ./
RUN go mod tidy

# Copy the application files
COPY ../ ./

# Build the Go application
RUN go build -o go-cdn main.go

# Use a minimal base image for the final container
FROM alpine:3.21

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/go-cdn /app/go-cdn

# Copy static files from the host (ensure /static exists)
COPY ../static /app/static

# Expose port 8080
EXPOSE 8080

# Start the application
CMD ["/app/go-cdn"]
