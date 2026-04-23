## Logi Backend

### Local Run
1. Update `configs/config.yaml` (or use env vars prefixed with `LOGI_`).
2. Build:
   ```bash
   go build -o logi ./cmd/backend
   ```
3. Run:
   ```bash
   ./logi
   ```

### Production-Safe Environment Variables
- `LOGI_ENVIRONMENT=production`
- `LOGI_SERVER_ADDRESS=:8080`
- `LOGI_MONGO_URI=<mongodb-uri>`
- `LOGI_JWT_SECRET=<32+ char random secret>`
- `LOGI_JWT_EXPIRATION_HOURS=72`
- `LOGI_MESSAGING_TYPE=websocket|nats`
- `LOGI_NATS_URL=nats://localhost:4222`
- `LOGI_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com`
- `LOGI_ENABLE_TEST_ROUTES=false`
- `LOGI_DB_OPERATION_TIMEOUT_SECONDS=5`

### Operational Endpoints
- `GET /healthz`
- `GET /readyz`
