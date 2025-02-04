I'll help you create a comprehensive README.md file:

```markdown:/Users/wanjohi/development/go/src/github.com/grocery-service/README.md
# Grocery Service

A modern, scalable grocery management service built with Go.

## Features

- 🛒 Product Management
- 👥 Customer Management
- 📦 Order Processing
- 📱 SMS Notifications (Africa's Talking)
- 📧 Email Notifications
- 🔐 JWT Authentication
- 📊 Category Management

## Tech Stack

- Go 1.21
- PostgreSQL 15
- Docker & Docker Compose
- JWT Authentication
- Africa's Talking SMS
- SMTP Email

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15
- Docker and Docker Compose
- Make

### Development Setup

1. Clone the repository
```bash
git clone https://github.com/yourusername/grocery-service.git
cd grocery-service
```

2. Set up environment variables
```bash
cp .env.example .env
```

3. Start services
```bash
make dev
```

4. Run migrations
```bash
make migrate-up
```

## Available Make Commands

```bash
make dev              # Start development environment
make migrate-up       # Run database migrations
make migrate-down     # Rollback migrations
make migrate-create   # Create new migration
make docker-build    # Build Docker images
make docker-up       # Start Docker containers
make docker-down     # Stop Docker containers
```

## API Documentation

### Authentication

All endpoints except `/auth/login` require JWT authentication:
```http
Authorization: Bearer <token>
```

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### Auth
- `POST /auth/login` - Login
- `POST /auth/register` - Register new user

#### Products
- `GET /products` - List products
- `GET /products/{id}` - Get product
- `POST /products` - Create product
- `PUT /products/{id}` - Update product
- `DELETE /products/{id}` - Delete product

#### Customers
- `GET /customers` - List customers
- `GET /customers/{id}` - Get customer
- `POST /customers` - Create customer
- `PUT /customers/{id}` - Update customer
- `DELETE /customers/{id}` - Delete customer

#### Orders
- `GET /orders` - List orders
- `GET /orders/{id}` - Get order
- `POST /orders` - Create order
- `PUT /orders/{id}/status` - Update order status

#### Categories
- `GET /categories` - List categories
- `POST /categories` - Create category
- `PUT /categories/{id}` - Update category
- `DELETE /categories/{id}` - Delete category

## Configuration

### Environment Variables

```env
# Server
SERVER_PORT=8080
SERVER_BASE_URL=http://localhost:8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=grocery_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key
JWT_ISSUER=grocery-service

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@grocery.com
SMTP_FROM_NAME=Grocery Service

# SMS (Africa's Talking)
SMS_API_KEY=your-api-key
SMS_USERNAME=your-username
SMS_SENDER_ID=GROCERY
SMS_ENVIRONMENT=sandbox
```

## Project Structure

```
.
├── cmd/
│   └── api/                 # Application entrypoint
├── internal/
│   ├── api/                # API handlers and routes
│   ├── config/             # Configuration
│   ├── domain/             # Domain models
│   ├── repository/         # Data access
│   ├── service/            # Business logic
│   └── utils/              # Utilities
├── migrations/             # Database migrations
├── .env.example           # Environment template
├── docker-compose.yml     # Docker compose config
├── Dockerfile             # Docker build file
├── go.mod                 # Go modules
└── Makefile              # Build commands
```

## Deployment

### Using Docker

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Manual Deployment

1. Build the binary:
```bash
go build -o grocery-service ./cmd/api
```

2. Run migrations:
```bash
make migrate-up
```

3. Start the service:
```bash
./grocery-service
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License.