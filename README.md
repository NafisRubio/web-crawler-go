# Web Crawler Go

A modern web crawler application built with Go 1.24 that fetches and parses product information from various websites.

## Features

- Fetch HTML content from URLs
- Parse product information using specialized providers
- Modular architecture with hexagonal design
- Redis integration for caching

## Project Structure

```
├── cmd/
│   └── server/         # Application entry points
├── internal/
│   ├── adapters/       # External adapters implementation
│   └── core/
│       ├── domain/     # Domain models
│       ├── ports/      # Interface definitions
│       └── services/   # Business logic implementation
```

## Requirements

- Go 1.24 or higher
- Redis (optional, for caching)

## Dependencies

- github.com/go-redis/redis/v8 - Redis client
- github.com/playwright-community/playwright-go - Web automation
- github.com/deckarep/golang-set/v2 - Set implementation

## Getting Started

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/web-crawler-go.git
cd web-crawler-go

# Install dependencies
go mod download
```

### Running the Application

```bash
go run cmd/server/main.go
```

## Usage

The web crawler service exposes endpoints to fetch product information from supported websites:

```
POST /api/v1/product
Body: {"url": "https://example.com/product/123"}
```

## Adding New Providers

To add support for a new website, implement the provider interface in the internal/adapters directory and register it with the service.

## License

MIT
