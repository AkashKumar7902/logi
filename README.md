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
- `LOGI_ENABLE_ADMIN_BOOTSTRAP=false`
- `LOGI_ADMIN_BOOTSTRAP_SECRET=<32+ char one-time bootstrap secret>`
- `LOGI_ENABLE_TEST_ROUTES=false`
- `LOGI_DB_OPERATION_TIMEOUT_SECONDS=5`

### Render Deploy
Render injects a `PORT` environment variable for web services, and the backend now uses it automatically if `LOGI_SERVER_ADDRESS` is not set. The backend also accepts these standard cloud aliases:

- `PORT`
- `MONGODB_URI` or `MONGO_URI`
- `JWT_SECRET`
- `ALLOWED_ORIGINS`

Minimum Render env setup:

```env
LOGI_ENVIRONMENT=production
MONGODB_URI=<your MongoDB Atlas or managed MongoDB URI>
JWT_SECRET=<32+ char random secret>
ALLOWED_ORIGINS=https://your-frontend.onrender.com
```

If `MONGODB_URI` is missing, or still points to `localhost`, startup now fails fast with a config error instead of timing out against `localhost:27017`.

### One-Time Admin Bootstrap
Public admin signup has been removed. To create the first admin account, temporarily enable the internal bootstrap route:

```env
LOGI_ENABLE_ADMIN_BOOTSTRAP=true
LOGI_ADMIN_BOOTSTRAP_SECRET=<32+ char random secret>
```

Then call:

```bash
curl -X POST http://localhost:8080/internal/bootstrap/admin \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Bootstrap-Secret: <same secret>' \
  -d '{"name":"Admin","email":"admin@example.com","password":"strong-password"}'
```

This route only works before the first admin exists. Disable `LOGI_ENABLE_ADMIN_BOOTSTRAP` immediately after bootstrap.

### Operational Endpoints
- `GET /healthz`
- `GET /readyz`
