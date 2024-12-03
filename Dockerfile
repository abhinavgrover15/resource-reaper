# Build stage
FROM golang:1.22.5-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o resource-reaper ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/resource-reaper .

# Set the binary as the entrypoint
ENTRYPOINT ["./resource-reaper"]
# Default command arguments can be overridden
CMD []
