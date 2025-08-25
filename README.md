# Web Crawler Go

A modern web crawler built with Go 1.24 that discovers and parses product information from supported e-commerce platforms, stores results in MongoDB, caches fetches in Redis, and streams real-time crawl updates to clients via Server-Sent Events (SSE).

## Features

- Extensible provider system (Shopify, Shopline; easy to add more)
- Product parsing with variants, images, and pricing
- Redis caching to reduce duplicate HTTP fetches
- MongoDB persistence and paginated querying
- Hexagonal architecture (ports/adapters) for clear separation of concerns
- HTTP API and SSE endpoints for real-time updates

## Current Project Structure

```
├── LICENSE
├── README.md
├── cmd/
│   └── server/
│       └── main.go
├── go.mod
├── go.sum
├── internal/
│   ├── adapters/
│   │   ├── primary/
│   │   │   └── http/
│   │   │       ├── crawler_handler.go
│   │   │       ├── middleware.go
│   │   │       ├── models.go
│   │   │       ├── product_handler.go
│   │   │       ├── response.go
│   │   │       ├── router.go
│   │   │       └── sse_handler.go
│   │   └── secondary/
│   │       ├── cache/
│   │       │   └── redis.go
│   │       ├── fetcher/
│   │       │   └── http.go
│   │       ├── providers/
│   │       │   ├── shopify/
│   │       │   │   └── parser.go
│   │       │   └── shopline/
│   │       │       ├── parser.go
│   │       │       └── types.go
│   │       └── repository/
│   │           └── mongodb.go
│   └── core/
│       ├── domain/
│       │   └── product.go
│       ├── ports/
│       │   ├── cache.go
│       │   ├── logger.go
│       │   └── ports.go
│       └── services/
│           ├── loggerservice/
│           │   └── logger.go
│           ├── productservice.go
│           └── sseservice.go
├── test_sse.html
└── test_sse_integration.go
```

## Requirements

- Go 1.24 or higher
- Running MongoDB and Redis instances (local or remote)

## Dependencies

- github.com/joho/godotenv — load env vars
- github.com/redis/go-redis/v9 — Redis client
- go.mongodb.org/mongo-driver/v2 — MongoDB driver
- golang.org/x/net — network utilities

## Getting Started

### 1) Install dependencies

```bash
go mod download
```

### 2) Configure environment

The server loads environment from .env.development by default (see cmd/server/main.go). Create that file in the project root and set the following:

```
# Database
MONGODB_URI=mongodb://localhost:27017

# Redis
REDIS_HOST=localhost:6379
REDIS_PASSWORD=

# HTTP server
PORT=8080
```

Adjust values if you use cloud providers or different ports.

### 3) Run the server

```bash
go run cmd/server/main.go
```

## API

Base path: /api/v1

- Crawl a domain
  - Method: GET
  - Path: /api/v1/crawl?domain_name=<domain>
  - Description: Crawls the given domain (e.g., example.com), parses products using the appropriate provider, stores them, and returns a count of saved products.
  - Response: { "status": "success", "message": "Domain crawled successfully", "data": { "productsCount": <int> } }

- List products by domain (paginated)
  - Method: GET
  - Path: /api/v1/products?domain_name=<domain>&page=<n>&page_size=<n>
  - Description: Returns products already stored for the given domain with pagination metadata.
  - Response: { "status": "success", "data": [ ...products ], "pagination": { page, page_size, total_items, total_pages, next_page, prev_page } }

- SSE stream
  - Method: GET
  - Path: /api/v1/sse?client_id=<optional>
  - Description: Establishes an SSE connection for real-time server messages (e.g., crawl progress heartbeats).

- SSE status
  - Method: GET
  - Path: /api/v1/sse/status
  - Description: Returns current SSE service status, including number of connected clients.

## Testing SSE locally

- test_sse.html: simple HTML page to connect to the SSE endpoint. Open it in a browser while the server is running.
- test_sse_integration.go: integration helper for validating SSE behavior.

## Adding new providers

Implement the ProductProvider in internal/core/ports and add your provider under internal/adapters/secondary/providers/<provider>. Then register it in cmd/server/main.go via providerRegistry["host"] = yourProvider.

## License

This project is licensed under the MIT License. See the LICENSE file for details.