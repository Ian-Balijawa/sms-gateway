# SMS Gateway API

A robust, high-performance SMS Gateway API built with Go that allows multiple client applications to send SMS messages through a unified API with API key/secret authentication.

## Features

- **Multi-Client Support**: Multiple client applications can register and use the API with unique API keys and secrets
- **API Key Authentication**: Secure API key/secret authentication for all SMS endpoints
- **Rate Limiting**: Configurable rate limits per client (daily, monthly, and per-second)
- **Bulk SMS Support**: Send single or bulk SMS messages
- **SMS Logging**: Complete logging of all SMS transactions with status tracking
- **Usage Statistics**: Track and monitor client usage statistics
- **Admin Panel**: Admin endpoints for managing clients and resetting usage
- **High Performance**: Built to handle multiple requests per second efficiently
- **Database Support**: Supports both SQLite (development) and PostgreSQL (production)

## Architecture

The API is structured as follows:

```
sms-gateway/
├── main.go              # Application entry point
├── config/              # Configuration management
├── models/              # Database models
├── database/            # Database initialization
├── handlers/            # HTTP request handlers
├── middleware/          # Authentication and authorization
├── service/             # SMS provider integration
└── utils/               # Utility functions
```

## Prerequisites

- Go 1.21 or higher installed
- egosms.co account credentials
- PostgreSQL (optional, for production) or SQLite (default, for development)

## Quick Start

### 1. Install Dependencies

```bash
go mod download
```

### 2. Configure Environment

```bash
cp env.template .env
# Edit .env with your egosms.co credentials
```

Fill in your `.env` file with:
- `SMS_USERNAME`: Your egosms.co username
- `SMS_PASSWORD`: Your egosms.co password
- `SMS_SENDER_ID`: Your default sender ID
- `ADMIN_PASSWORD`: Change from default for security
- `JWT_SECRET`: Generate with `openssl rand -hex 32`

### 3. Run the Server

```bash
go run main.go
# Or use Makefile:
make run
```

The API will start on `http://localhost:8080` by default.

### 4. Create Your First Client

Using curl:
```bash
curl -X POST http://localhost:8080/api/v1/admin/clients \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My App",
    "email": "myapp@example.com"
  }'
```

**Important**: Save the `api_key` and `api_secret` from the response - you'll need them to send SMS.

### 5. Send Your First SMS

```bash
curl -X POST http://localhost:8080/api/v1/sms/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "X-API-Secret: YOUR_API_SECRET" \
  -d '{
    "number": "+256701234567",
    "message": "Hello from SMS Gateway!"
  }'
```

## Installation

For detailed installation instructions, see the [Quick Start](#quick-start) section above.

### Building

```bash
go build -o sms-gateway main.go
```

### Using Makefile

```bash
make run      # Run the application
make build    # Build the binary
make test     # Run tests
make clean    # Clean build artifacts
```

## API Endpoints

### Health Check

```http
GET /health
```

### SMS Endpoints (Require API Key Authentication)

All SMS endpoints require the following headers:
- `X-API-Key`: Your API key
- `X-API-Secret`: Your API secret

#### Send Single SMS

```http
POST /api/v1/sms/send
Content-Type: application/json

{
  "number": "+256701234567",
  "message": "Hello, this is a test message",
  "senderid": "MyApp",
  "priority": "1"
}
```

#### Send Bulk SMS

```http
POST /api/v1/sms/send/bulk
Content-Type: application/json

{
  "messages": [
    {
      "number": "+256701234567",
      "message": "Hello User 1"
    },
    {
      "number": "+256709876543",
      "message": "Hello User 2"
    }
  ]
}
```

#### Get SMS Logs

```http
GET /api/v1/sms/logs?limit=50&offset=0&status=sent
```

Query Parameters:
- `limit`: Number of logs to return (default: 50)
- `offset`: Pagination offset (default: 0)
- `status`: Filter by status (pending, sent, failed)

#### Get Statistics

```http
GET /api/v1/sms/stats
```

Returns usage statistics for the authenticated client.

### Admin Endpoints (Require Basic Auth)

Admin endpoints require Basic Authentication. Set credentials via `ADMIN_USER` and `ADMIN_PASSWORD` environment variables.

#### Create Client

```http
POST /api/v1/admin/clients
Authorization: Basic <base64(username:password)>
Content-Type: application/json

{
  "name": "Client Name",
  "email": "client@example.com",
  "rate_limit": 100,
  "daily_limit": 10000,
  "monthly_limit": 300000
}
```

#### List Clients

```http
GET /api/v1/admin/clients?is_active=true
```

#### Update Client

```http
PUT /api/v1/admin/clients/{client_id}
Authorization: Basic <base64(username:password)>
Content-Type: application/json

{
  "name": "Updated Name",
  "is_active": true,
  "rate_limit": 200,
  "daily_limit": 20000
}
```

#### Reset Client Usage

```http
POST /api/v1/admin/clients/{client_id}/reset
Authorization: Basic <base64(username:password)>
```

## Example Usage

### Using cURL

```bash
# Send a single SMS
curl -X POST http://localhost:8080/api/v1/sms/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "X-API-Secret: your-api-secret" \
  -d '{
    "number": "+256701234567",
    "message": "Hello from SMS Gateway!"
  }'
```

### Using Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func main() {
    url := "http://localhost:8080/api/v1/sms/send"
    
    payload := map[string]string{
        "number":  "+256701234567",
        "message": "Hello from Go!",
    }
    
    jsonData, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", "your-api-key")
    req.Header.Set("X-API-Secret", "your-api-secret")
    
    client := &http.Client{}
    resp, _ := client.Do(req)
    defer resp.Body.Close()
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server host address | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8080` |
| `DB_TYPE` | Database type (`sqlite` or `postgres`) | `sqlite` |
| `DB_HOST` | Database host (PostgreSQL) | `localhost` |
| `DB_PORT` | Database port (PostgreSQL) | `5432` |
| `DB_USER` | Database user (PostgreSQL) | `postgres` |
| `DB_PASSWORD` | Database password (PostgreSQL) | `postgres` |
| `DB_NAME` | Database name | `sms_gateway` |
| `SMS_USERNAME` | egosms.co username | - |
| `SMS_PASSWORD` | egosms.co password | - |
| `SMS_SENDER_ID` | Default sender ID | - |
| `SMS_SANDBOX_MODE` | Use sandbox mode | `true` |
| `RATE_LIMIT_RPS` | Global rate limit (requests per second) | `100` |
| `ADMIN_USER` | Admin username | `admin` |
| `ADMIN_PASSWORD` | Admin password | `admin` |

## Production Deployment

For production deployments, consider:

1. **Use PostgreSQL** instead of SQLite:
   - Set `DB_TYPE=postgres` in `.env`
   - Configure PostgreSQL connection details

2. **Change Admin Credentials**:
   - Set strong `ADMIN_USER` and `ADMIN_PASSWORD` in `.env`

3. **Set Strong JWT Secret**:
   - Generate a secure random string for `JWT_SECRET`:
     ```bash
     openssl rand -hex 32
     ```

4. **Configure CORS**:
   - Update CORS settings in `main.go` to allow only your domains

5. **Use Reverse Proxy**:
   - Deploy behind nginx or similar for SSL termination

6. **Monitor Logs**:
   - Set up log aggregation for production monitoring

## Performance Considerations

- The API is designed to handle multiple concurrent requests efficiently
- Database connections are managed by GORM connection pooling
- For production deployments with high traffic, consider:
  - Using PostgreSQL instead of SQLite
  - Implementing Redis for distributed rate limiting
  - Adding request queuing for bulk SMS operations
  - Using a reverse proxy (nginx) for load balancing

### Testing High Load

The API is designed to handle multiple requests per second. To test:

```bash
# Install hey (HTTP load testing tool)
go install github.com/rakyll/hey@latest

# Test with 100 concurrent requests, 1000 total
hey -n 1000 -c 100 \
  -H "X-API-Key: YOUR_KEY" \
  -H "X-API-Secret: YOUR_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"number":"+256701234567","message":"Test"}' \
  http://localhost:8080/api/v1/sms/send
```

## Security

- API secrets are hashed using bcrypt before storage
- Rate limiting prevents abuse
- Client status can be toggled to disable access
- All SMS transactions are logged for audit purposes

## Database Schema

### APIClient
- Stores client information and credentials
- Tracks usage limits and current usage
- Manages client status (active/inactive)

### SMSLog
- Logs every SMS transaction
- Stores recipient, message, status, and provider responses
- Links to client for tracking

## Development

### Running Tests

```bash
go test ./...
```

### Development Mode

```bash
# Run with debug logging
GIN_MODE=debug go run main.go
# Or use Makefile
make dev
```

## License

[Your License Here]

## Support

For issues and questions, please open an issue in the repository.

