# GoREST Metrics Plugin

[![CI](https://github.com/nicolasbonnici/gorest-metrics/actions/workflows/ci.yml/badge.svg)](https://github.com/nicolasbonnici/gorest-metrics/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicolasbonnici/gorest-metrics)](https://goreportcard.com/report/github.com/nicolasbonnici/gorest-metrics)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A production-ready GoREST plugin for tracking integer metrics on polymorphic resources. Track view counts, downloads, likes, or any custom metric with historical data and efficient querying.

## Features

- **Polymorphic Metrics**: Track metrics for any resource type (posts, users, products, etc.)
- **Flexible Values**: Support for both positive and negative integers (deltas, adjustments)
- **Unique Constraints**: Enforces one metric per (resource, resource_id, name) combination
- **Historical Tracking**: Automatic timestamp tracking with `created_at`
- **Advanced Filtering**: Filter by resource type, ID, name, or value ranges
- **Multi-Database**: Full support for PostgreSQL, MySQL, and SQLite
- **Efficient Indexing**: Optimized composite indexes for fast queries
- **Type Safety**: Generic CRUD operations with compile-time type checking

## Installation

```bash
go get github.com/nicolasbonnici/gorest-metrics
```


## Development Environment

To set up your development environment:

```bash
make install
```

This will:
- Install Go dependencies
- Install development tools (golangci-lint)
- Set up git hooks (pre-commit linting and tests)

## Quick Start

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/nicolasbonnici/gorest"
    "github.com/nicolasbonnici/gorest-metrics"
)

func main() {
    app := gorest.New(gorest.Config{
        Database: db,
        Plugins: []plugin.Plugin{
            metrics.NewPlugin(),
        },
        PluginConfigs: map[string]map[string]interface{}{
            "metrics": {
                "allowed_types": []interface{}{"post", "user", "product"},
                "max_name_length": 255,
                "only_positive_values": false,
                "pagination_limit": 50,
                "max_pagination_limit": 200,
            },
        },
    })

    app.Listen(":8080")
}
```

## Configuration

### Config Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `allowed_types` | `[]string` | `["post"]` | Resource types that can have metrics |
| `max_key_length` | `int` | `255` | Maximum length for metric keys (1-255) |
| `only_positive_values` | `bool` | `false` | Restrict values to positive integers only |
| `pagination_limit` | `int` | `50` | Default page size for list queries |
| `max_pagination_limit` | `int` | `200` | Maximum allowed page size (1-1000) |

### Example Configuration (YAML)

```yaml
plugins:
  metrics:
    allowed_types:
      - post
      - user
      - product
    max_key_length: 255
    only_positive_values: false
    pagination_limit: 50
    max_pagination_limit: 200
```

## Database Schema

The plugin automatically creates a `metrics` table with the following structure:

```sql
CREATE TABLE metrics (
    id UUID PRIMARY KEY,
    resource TEXT NOT NULL,              -- Resource type (post, user, etc.)
    resource_id UUID NOT NULL,            -- Foreign key to resource
    name VARCHAR(255) NOT NULL,            -- Metric name (views, etc.)
    value INTEGER NOT NULL DEFAULT 0,     -- Metric value (supports negative)
    created_at TIMESTAMP NOT NULL,        -- Last update timestamp
    UNIQUE (resource, resource_id, name)   -- One metric per resource/name
);

-- Composite indexes for performance
CREATE INDEX idx_metrics_resource ON metrics(resource, resource_id, name);
CREATE INDEX idx_metrics_name ON metrics(name, created_at);
CREATE INDEX idx_metrics_resource_id ON metrics(resource_id);
```

## API Endpoints

### List Metrics

Get all metrics with filtering, pagination, and ordering.

```http
GET /metrics?resource=post&resourceId={uuid}&name=views&limit=50&page=1
```

**Query Parameters:**

- `resource` - Filter by resource type
- `resourceId` - Filter by specific resource UUID
- `name` - Filter by metric  name
- `value[eq]`, `value[gt]`, `value[gte]`, `value[lt]`, `value[lte]` - Filter by value ranges
- `limit` - Results per page (default: 50, max: 200)
- `page` - Page number (default: 1)
- `order` - Sort order (e.g., `-value`, `createdAt`, `-name`)
- `count` - Include total count (default: true)

**Example Response:**

```json
{
  "@context": "/api/contexts/Metric",
  "@id": "/api/metrics",
  "@type": "hydra:Collection",
  "hydra:member": [
    {
      "@id": "/api/metrics/123e4567-e89b-12d3-a456-426614174000",
      "@type": "Metric",
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "resource": "post",
      "resourceId": "550e8400-e29b-41d4-a716-446655440000",
      "name": "views",
      "value": 1250,
      "createdAt": "2026-02-08T10:30:00Z"
    }
  ],
  "hydra:totalItems": 1,
  "hydra:view": {
    "@id": "/api/metrics?page=1",
    "@type": "hydra:PartialCollectionView",
    "hydra:first": "/api/metrics?page=1",
    "hydra:last": "/api/metrics?page=1"
  }
}
```

### Get Metric by ID

```http
GET /metrics/{id}
```

**Example Response:**

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "resource": "post",
  "resourceId": "550e8400-e29b-41d4-a716-446655440000",
  "name": "views",
  "value": 1250,
  "createdAt": "2026-02-08T10:30:00Z"
}
```

### Create Metric

```http
POST /metrics
Content-Type: application/json

{
  "resource": "post",
  "resourceId": "550e8400-e29b-41d4-a716-446655440000",
  "name": "views",
  "value": 1
}
```

**Response:** `201 Created` with created metric object

**Note:** Creating a metric with duplicate (resource, resourceId, name) will fail with `500 Internal Server Error`. Consider using Update endpoint to modify existing metrics.

### Update Metric

Update only the value of an existing metric.

```http
PUT /metrics/{id}
Content-Type: application/json

{
  "value": 1251
}
```

**Response:** `200 OK` with updated metric object

### Delete Metric

```http
DELETE /metrics/{id}
```

**Response:** `204 No Content`

## Use Cases

### View Counter

```bash
# Increment view count for a blog post
curl -X POST http://localhost:8080/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "resource": "post",
    "resourceId": "550e8400-e29b-41d4-a716-446655440000",
    "name": "views",
    "value": 1
  }'

# Get current view count
curl "http://localhost:8080/metrics?resource=post&resourceId=550e8400-e29b-41d4-a716-446655440000&name=views"

# Update view count (after fetching metric ID)
curl -X PUT http://localhost:8080/metrics/{id} \
  -H "Content-Type: application/json" \
  -d '{"value": 125}'
```

### Download Tracker

```bash
# Track product downloads
curl -X POST http://localhost:8080/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "resource": "product",
    "resourceId": "650e8400-e29b-41d4-a716-446655440000",
    "name": "download_count",
    "value": 1
  }'

# Get top 10 most downloaded products
curl "http://localhost:8080/metrics?name=download_count&order=-value&limit=10"
```

### Reputation System

```bash
# Track user reputation (supports negative values)
curl -X POST http://localhost:8080/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "resource": "user",
    "resourceId": "750e8400-e29b-41d4-a716-446655440000",
    "name": "reputation",
    "value": 100
  }'

# Decrease reputation
curl -X PUT http://localhost:8080/metrics/{id} \
  -H "Content-Type: application/json" \
  -d '{"value": 90}'
```

### Analytics Dashboard

```bash
# Get all metrics for a specific post
curl "http://localhost:8080/metrics?resource=post&resourceId=550e8400-e29b-41d4-a716-446655440000"

# Get metrics with high values (engagement threshold)
curl "http://localhost:8080/metrics?value[gte]=1000&order=-value"

# Get recently updated metrics
curl "http://localhost:8080/metrics?order=-createdAt&limit=20"
```

## Development

### Prerequisites

- Go 1.25.1+
- golangci-lint (for linting)

### Setup

```bash
# Clone the repository
git clone https://github.com/nicolasbonnici/gorest-metrics.git
cd gorest-metrics

# Install dependencies
go mod download

# Install development tools
make install
```

### Available Make Targets

```bash
make help              # Show all available targets
make test              # Run tests with race detection and coverage
make coverage          # Generate HTML coverage report
make lint              # Run golangci-lint
make lint-fix          # Run linter with auto-fix
make build             # Build verification
make clean             # Clean caches and build artifacts
```

### Running Tests

```bash
# Run all tests with coverage
make test

# Generate coverage report
make coverage
# Open coverage.html in browser

# Run with verbose output
go test -v -race ./...
```

### Git Hooks

Install pre-commit hooks to ensure code quality:

```bash
./.githooks/install.sh
```

The pre-commit hook runs:
- `make lint` - Code linting
- `make test` - Full test suite

To bypass hooks (not recommended):
```bash
git commit --no-verify
```

## Security

- **Input Validation**: All requests validated before database operations
- **Resource Type Whitelist**: Only configured resource types allowed
- **UUID Validation**: Strict UUID format enforcement for resource IDs
- **name Length Limits**: Configurable maximum name length (1-255)
- **Value Constraints**: Optional positive-only value enforcement
- **Filter Limits**: Maximum 50 values per filter field to prevent abuse

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

- Issues: [GitHub Issues](https://github.com/nicolasbonnici/gorest-metrics/issues)
- Documentation: [GoREST Documentation](https://github.com/nicolasbonnici/gorest)

## Related Projects

- [gorest](https://github.com/nicolasbonnici/gorest) - Core GoREST framework
- [gorest-likeable](https://github.com/nicolasbonnici/gorest-likeable) - Like/unlike functionality
- [gorest-commentable](https://github.com/nicolasbonnici/gorest-commentable) - Comment system
- [gorest-translatable](https://github.com/nicolasbonnici/gorest-translatable) - Multi-language support
