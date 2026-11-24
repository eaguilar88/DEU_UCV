# Stage 1: Build
FROM golang:1.27.4-alpine3.24 AS builder

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
RUN go build -o main cmd/deu/main.go && \
    go build -o notifier ./cmd/notifier && \
    go build -o backup ./cmd/backup

# Stage 2: Final Image
FROM alpine:3.22

WORKDIR /app

# pg_dump for the backup service (cmd/backup); same major version as the db service (postgres:16)
RUN apk add --no-cache postgresql16-client

# Copy only the necessary files from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/notifier .
COPY --from=builder /app/backup .

# Expose port 8081
EXPOSE 8081

# Set the entrypoint
ENTRYPOINT ["./main"]
