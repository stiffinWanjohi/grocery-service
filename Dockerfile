# Build stage
FROM golang:1.23.5-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application and migration tool
RUN CGO_ENABLED=1 GOOS=linux go build -a -o grocery-service ./cmd/api
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN cp $(which migrate) .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata postgresql-client

# Copy binaries, config and migrations
COPY --from=builder /app/grocery-service .
COPY --from=builder /app/migrate .
COPY --from=builder /app/.env .env
COPY --from=builder /app/migrations ./migrations

ENV TZ=UTC

EXPOSE 8080

# Entrypoint script
CMD ["sh", "-c", "\
    until PGPASSWORD=$DB_PASSWORD psql -h \"$DB_HOST\" -U \"$DB_USER\" -d \"$DB_NAME\" -c '\\q'; do \
    echo 'Waiting for postgres...'; \
    sleep 1; \
    done; \
    echo 'PostgreSQL is up - running migrations...'; \
    ./migrate -path migrations -database \"$DB_URL\" up && \
    echo 'Starting application...' && \
    ./grocery-service \
    "]