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
RUN go build -o go-cdn *.go

# Use a minimal base image for the final container
FROM alpine:3.21

# Create a new user and group
RUN addgroup -S gocdn-group && adduser -S gocdn-user -G gocdn-group

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/go-cdn /app/go-cdn

# Copy static files from the host (ensure /static exists)
COPY ../static /app/static

# Change ownership of the application files to the new user
RUN chown -R gocdn-user:gocdn-group /app

# Switch to the new user
USER gocdn-user

# Expose port 8080
EXPOSE 8080

# Add HEALTHCHECK instruction
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["wget", "--spider", "http://localhost:8080/health"]

# Start the application
CMD ["/app/go-cdn"]
