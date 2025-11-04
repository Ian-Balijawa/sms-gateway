# Quick Start Guide

## Prerequisites

- Go 1.21 or higher installed
- egosms.co account credentials

## Setup Steps

1. **Install Dependencies**
```bash
go mod download
```

2. **Configure Environment**
```bash
cp .env.example .env
# Edit .env with your egosms.co credentials
```

3. **Run the Server**
```bash
go run main.go
# Or use Makefile:
make run
```

4. **Create Your First Client**

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

Save the `api_key` and `api_secret` from the response - you'll need them to send SMS.

5. **Send Your First SMS**

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

## Production Deployment

For production, consider:

1. **Use PostgreSQL** instead of SQLite:
   - Set `DB_TYPE=postgres` in `.env`
   - Configure PostgreSQL connection details

2. **Change Admin Credentials**:
   - Set strong `ADMIN_USER` and `ADMIN_PASSWORD`

3. **Set Strong JWT Secret**:
   - Generate a secure random string for `JWT_SECRET`

4. **Configure CORS**:
   - Update CORS settings in `main.go` to allow only your domains

5. **Use Reverse Proxy**:
   - Deploy behind nginx or similar for SSL termination

6. **Monitor Logs**:
   - Set up log aggregation for production monitoring

## Testing High Load

The API is designed to handle multiple requests per second. To test:

```bash
# Install hey (HTTP load testing tool)
go install github.com/rakyll/hey@latest

# Test with 100 concurrent requests, 1000 total
hey -n 1000 -c 100 -H "X-API-Key: YOUR_KEY" -H "X-API-Secret: YOUR_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"number":"+256701234567","message":"Test"}' \
  http://localhost:8080/api/v1/sms/send
```

