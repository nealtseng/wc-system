# Stage 1: Build the Go binary
FROM golang:1.26-alpine AS builder
WORKDIR /app

# Download dependencies separately so Docker layer caching avoids re-downloading
# on code-only changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically-linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./main.go

# Stage 2: Minimal runtime image
FROM alpine:3.20
WORKDIR /app

# Copy the compiled binary from the builder stage.
COPY --from=builder /app/server .

# Copy SQL migration files so the server can apply them at startup.
COPY db/migrations ./db/migrations

# Optional manual xG seed CSV (used when FBref scrape is blocked).
COPY data ./data

EXPOSE 8080
CMD ["./server"]
