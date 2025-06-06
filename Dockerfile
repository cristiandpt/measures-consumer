# Stage 1: Builder (to compile the Go application)
FROM golang:1.24.2-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download -x

# Copy the entire project source code
COPY . .

# Build the Go application
RUN go build -o rabbitmq-consumer ./cmd/consumer/main.go

# Stage 2: Final Image (lightweight runtime image)
FROM alpine:latest

# Install necessary runtime dependencies (if any) - in this case, likely none for a simple Go app
# RUN apk add --no-cache <your-runtime-dependencies>

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/rabbitmq-consumer .

# Command to run the application
CMD ["./rabbitmq-consumer"]
