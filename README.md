# Caching Proxy

https://roadmap.sh/projects/caching-server

A simple CLI tool written in Go that starts a caching proxy server. The proxy forwards incoming HTTP requests to an origin server, caches the responses in memory, and serves subsequent requests from the cache to improve performance. The response headers include an `X-Cache` field indicating whether the response was served from cache (`HIT`) or from the origin server (`MISS`).

## Features

- **CLI Interface**: Start the proxy using command-line flags.
- **Caching**: Uses an in-memory cache powered by [go-cache](https://github.com/patrickmn/go-cache) with a default TTL of 5 minutes.
- **Reverse Proxy**: Forwards requests to the specified origin server and caches the responses.
- **Cache Clearing**: Clear the in-memory cache via the `--clear-cache` flag.

## Requirements

- Go 1.18 or later

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/nabobery/caching-proxy.git
   cd caching-proxy
   ```

2. Build the project:

```bash
   # For Linux/Mac
   go build -o caching-proxy

   # For Windows
   go build -o caching-proxy.exe main.go
```

## Usage

### Start the Proxy Server

To start the caching proxy server, specify the port and the origin URL:

```bash
    ./caching-proxy --port 3000 --origin http://dummyjson.com
```

This starts the proxy on port 3000. For example, if you access `http://localhost:3000/products`, the proxy will forward the request to `http://dummyjson.com/products`. The response will include the header `X-Cache: MISS` for the first request and `X-Cache: HIT` for subsequent requests to the same endpoint.

### Clear the Cache

To clear the in-memory cache, run:

```bash
    ./caching-proxy --clear-cache
```

This will flush the cache and output a confirmation message.

## Project

```bash
caching-proxy/
├── cache
│ └── cache.go # In-memory cache implementation using go-cache
├── cmd
│ └── root.go # CLI setup using Cobra
├── proxy
│ ├── proxy.go # Reverse proxy server implementation
│ └── caching_middleware.go # Custom transport for caching responses
├── main.go # Entry point of the application
├── go.mod # Module definition and dependencies
└── README.md # Project documentation
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License.

Happy coding!
