# Build stage
FROM golang:1.23.5-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application and seeder
RUN CGO_ENABLED=1 GOOS=linux go build -a -o app ./cmd/api && \
    CGO_ENABLED=1 GOOS=linux go build -a -o seeder ./cmd/seed

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies and golang-migrate
RUN apk add --no-cache ca-certificates tzdata postgresql-client curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/

# Copy application binary, seeder, and migrations
COPY --from=builder /build/app .
COPY --from=builder /build/seeder .
COPY --from=builder /build/migrations ./migrations

# Ensure executable permissions
RUN chmod +x /app/app && \
    chmod +x /app/seeder

# Set environment variables
ENV TZ=Africa/Nairobi

# Wait for PostgreSQL, run migrations, seed data, then start the application
CMD set -e; \
    echo "Waiting for PostgreSQL..."; \
    until PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -c '\q'; do \
    echo "PostgreSQL is unavailable - sleeping"; \
    sleep 1; \
    done; \
    echo "PostgreSQL is up - running migrations for main database"; \
    migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" up; \
    echo "Seeding main database..."; \
    ./seeder; \
    echo "Running migrations for test database"; \
    migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${TEST_DB_NAME}?sslmode=disable" up; \
    echo "Starting application"; \
    exec /app/app