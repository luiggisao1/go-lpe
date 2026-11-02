# Stage 1: Build static Go binary
FROM golang:alpine AS builder

WORKDIR /app

# Download dependencies (cached layer)
COPY go.mod ./
# COPY go.sum ./ # Uncomment once external packages are added
RUN go mod download

# Copy application source code
COPY . .

# Compile a fully static binary
# CGO_ENABLED=0 ensures pure Go without C library dependencies
# -ldflags="-w -s" strips debug symbols to minimize binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/server ./cmd/server

# Stage 2: Minimal runtime environment
FROM alpine:latest

WORKDIR /app

# Install CA certificates for secure HTTPS outgoing connections (Slack, Gemini, Notion)
RUN apk --no-cache add ca-certificates tzdata

# Create a non-privileged user for security best practices
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy only the compiled binary from the builder stage
COPY --from=builder /bin/server /app/server

# Expose HTTP service port
EXPOSE 8080

# Environment defaults
ENV PORT=8080
ENV ENV=production

# Run binary
CMD ["/app/server"]
