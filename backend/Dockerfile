# Stage 1: Build
FROM golang:1.26.1-alpine3.22 AS builder

# Install required packages and set the working directory
RUN apk add --no-cache git && \
    mkdir -p /app

WORKDIR /app

# Copy Go modules manifests first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the remaining project files
COPY . .

# Build the project
RUN go build -o main cmd/deu/main.go

# Stage 2: Final Image
FROM alpine:3.22

WORKDIR /app

# Copy only the necessary files from the builder stage
COPY --from=builder /app/main .

# Expose port 8081
EXPOSE 8081

# Set the entrypoint
ENTRYPOINT ["./main"]
