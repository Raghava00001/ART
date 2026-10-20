# Stage 1: Build the standalone binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and static web assets
COPY . .

# Build stripped, zero-dependency static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /art .

# Stage 2: Minimal, secure production container
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy compiled standalone binary from builder
COPY --from=builder /art /app/art

# Expose web atelier port
EXPOSE 8080

# Run in Web GUI mode by default, respecting dynamic $PORT if set
ENTRYPOINT ["/app/art", "-gui"]
