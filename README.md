# Web Crawler Go

A modern web crawler application built with Go 1.24 that fetches and parses product information from various websites.

## Features

- **Extensible Provider System:** Easily add new providers to support different e-commerce platforms like Shopify and Shopline.
- **Product Information Parsing:** Extracts detailed product information including name, price, images, and variants.
- **Caching Layer:** Utilizes Redis to cache fetched data, reducing redundant requests and improving performance.
- **Database Integration:** Stores parsed product data in MongoDB for persistence and querying.
- **Hexagonal Architecture:** A clean and modular design that separates core logic from external concerns, making the application easier to maintain and test.
- **RESTful API:** Provides a simple and intuitive API for interacting with the web crawler service.

## Project Structure

```
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── adapters/
│   │   ├── primary/
│   │   │   └── http/
│   │   │       ├── handler.go
│   │   │       ├── middleware.go
│   │   │       ├── response.go
│   │   │       └── router.go
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
│           ├── productservice.go
│           └── loggerservice/
│               └── logger.go
```

## Requirements

- Go 1.24 or higher
- Docker (for running MongoDB and Redis)
- GoLand (or any other Go-compatible IDE)

## Dependencies

- `github.com/joho/godotenv` - for managing environment variables
- `github.com/redis/go-redis/v9` - Redis client for Go
- `go.mongodb.org/mongo-driver/v2` - MongoDB driver for Go
- `golang.org/x/net` - for network-related functionalities

## Getting Started

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/yourusername/web-crawler-go.git
    cd web-crawler-go
    ```

2.  **Install dependencies:**

    ```bash
    go mod download
    ```

3.  **Set up environment variables:**

    Create a `.env` file in the root of the project and add the following variables:

    ```
    # Database Configuration
    MONGODB_URI=mongodb+srv://<user>:<password>@<cluster-uri>/<database>

    # Redis Configuration
    REDIS_HOST=localhost:6379
    REDIS_PASSWORD=

    # Server Configuration
    PORT=8080

    # Other configurations
    LOG_LEVEL=info
    ```

    Replace the placeholder values with your actual MongoDB connection details.

### Running the Application

1.  **Start the database and cache:**

    You can use Docker to easily run MongoDB and Redis:

    ```bash
    docker-compose up -d
    ```

2.  **Run the application:**

    ```bash
    go run cmd/server/main.go
    ```

## Usage

The web crawler service exposes the following endpoint to fetch product information:

### Fetch Product Information

- **Endpoint:** `POST /api/v1/product`
- **Description:** Fetches and parses product information from the given URL.
- **Body:**

  ```json
  {
    "url": "https://example.com/product/123"
  }
  ```

- **Response:**

  Returns a JSON object with the parsed product information.

## Adding New Providers

To add support for a new website, you need to implement the `Parser` interface from the `internal/core/ports` directory.

1.  **Create a new provider directory:**

    Create a new directory under `internal/adapters/secondary/providers` for the new provider (e.g., `internal/adapters/secondary/providers/newprovider`).

2.  **Implement the `Parser` interface:**

    Create a new Go file in the provider directory and implement the `Parse` method. This method will contain the logic for parsing the product information from the website's HTML.

3.  **Register the new provider:**

    Register the new provider in the `ProductService` so that it can be used by the application.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.