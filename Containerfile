# Use official Go image as base
FROM golang:alpine3.22 AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o md2notion cmd/notion-md/main.go

# Use minimal base image for final image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/md2notion .

# Make it executable
RUN chmod +x ./md2notion

# Set default entrypoint
ENTRYPOINT ["./md2notion"]
