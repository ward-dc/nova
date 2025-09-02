# Nova API

Solana wallet balance API with caching and authentication.

## Setup

```bash
go mod tidy
go run main.go
```

Requires MongoDB and optionally Redis/DragonflyDB for caching.

## Usage

```bash
curl -X POST http://localhost:8080/api/get-balance \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"wallets": ["wallet1", "wallet2"]}'
```

## Testing

```bash
go test ./test/ -v
```

## Architecture

- MongoDB for API key storage
- Redis + memory caching for performance (I decided to go with memory for token and rate limit caching as it reduces the amount of network calls. DragonflyDB is used for balance caching as this might be accessed by multiple services)
- Rate limiting
- Per-wallet mutexes prevent race conditions
- Rate limit on the max amount of wallets per request (Was not in requirements but I added it as it seems logical to limit the amount of wallets per request)
- API key caching (Was not in requirements but I added it as it reduces the amount of network calls)
- Dockerfile and GitHub Actions

