# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy everything
COPY . .

# Generate go.sum and download dependencies
RUN go mod tidy
RUN go mod download

# Build with CGO for SQLite
RUN CGO_ENABLED=1 go build -o /app/bot ./cmd/bot

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bot /app/bot

# Create data directory
RUN mkdir -p /data

EXPOSE 8080

CMD ["/app/bot"]
