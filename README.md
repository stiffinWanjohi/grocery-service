I'll help you update the README.md file with the project structure and testing information:

```markdown:/Users/wanjohi/development/go/src/github.com/grocery-service/README.md
# Grocery Service

A modern, scalable grocery management service built with Go.

## Features

- 🛒 Product Management
- 👥 Customer Management
- 📦 Order Processing
- 🔐 JWT Authentication
- 📊 Category Management
- 🔄 Real-time Stock Updates
- 📝 API Documentation with Swagger

## Tech Stack

- Go 1.21+
- PostgreSQL
- Chi Router
- JWT Authentication
- Swagger Documentation

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL
- Make

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/grocery-service.git
cd grocery-service
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
```

4. Run the application:
```bash
go run cmd/api/main.go
```

## API Documentation

The API documentation is available via Swagger UI at:
```
http://localhost:8080/swagger/index.html
```

## Project Structure

```
.
├── cmd/
│   └── api/                 # Application entrypoint
├── internal/
│   ├── api/
│   │   ├── handlers/       # HTTP handlers
│   │   └── middleware/     # HTTP middleware
│   ├── config/             # Configuration
│   ├── domain/             # Domain models
│   ├── repository/         # Data access layer
│   │   └── postgres/       # PostgreSQL implementations
│   ├── service/            # Business logic
│   └── utils/              # Utility packages
├── docs/                   # Documentation
│   └── swagger/            # Swagger files
├── tests/                  # Test utilities
│   └── mocks/             # Mock implementations
└── migrations/             # Database migrations
```

## Testing

### Running Tests

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Structure

Tests are organized following the same structure as the production code:

- Unit tests are placed next to the code they test
- Integration tests are in separate test files
- Mocks are generated using testify/mock

Example test file structure:
```
internal/
├── service/
│   ├── category.go
│   └── category_test.go
└── repository/
    ├── postgres/
    │   ├── category.go
    │   └── category_test.go
```

## API Endpoints

### Categories
- `GET /api/v1/categories` - List all categories
- `POST /api/v1/categories` - Create a new category
- `GET /api/v1/categories/{id}` - Get category by ID
- `PUT /api/v1/categories/{id}` - Update category
- `DELETE /api/v1/categories/{id}` - Delete category
- `GET /api/v1/categories/{id}/subcategories` - List subcategories

### Products
- `GET /api/v1/products` - List all products
- `POST /api/v1/products` - Create a new product
- `GET /api/v1/products/{id}` - Get product by ID
- `PUT /api/v1/products/{id}` - Update product
- `DELETE /api/v1/products/{id}` - Delete product
- `PUT /api/v1/products/{id}/stock` - Update product stock
- `GET /api/v1/products/category/{categoryID}` - List products by category

### Customers
- `GET /api/v1/customers` - List all customers
- `POST /api/v1/customers` - Create a new customer
- `GET /api/v1/customers/{id}` - Get customer by ID
- `PUT /api/v1/customers/{id}` - Update customer
- `DELETE /api/v1/customers/{id}` - Delete customer

### Orders
- `GET /api/v1/orders` - List all orders
- `POST /api/v1/orders` - Create a new order
- `GET /api/v1/orders/{id}` - Get order by ID
- `GET /api/v1/orders/customer/{customerID}` - List customer orders
- `PUT /api/v1/orders/{id}/status` - Update order status
- `POST /api/v1/orders/{id}/items` - Add order item
- `DELETE /api/v1/orders/{id}/items/{itemID}` - Remove order item

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
```