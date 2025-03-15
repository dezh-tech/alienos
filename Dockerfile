FROM golang:1.24.1 AS builder

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build a statically linked binary compatible with Alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o alienos

# Use a minimal base image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install CA certificates (if needed for HTTPS calls)
RUN apk add --no-cache ca-certificates

# Copy the built binary from the builder stage
COPY --from=builder /app/alienos .

# Ensure the binary is executable
RUN chmod +x alienos

# Expose the relay port
EXPOSE 7771

# Run the application
CMD ["./alienos"]
