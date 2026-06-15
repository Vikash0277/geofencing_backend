# Geofencing Backend — Setup Guide

## Prerequisites

- Go 1.25+
- PostgreSQL 15+ with PostGIS extension
- Docker & Docker Compose (optional)

## Local Setup

### 1. Clone and enter the backend directory

```bash
cd geofencing_backend
```

### 2. Configure environment

Copy `.env.dev` to `.env` and adjust values:

```bash
cp .env.dev .env
```

Key variables:

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `MIGRATION` | Set to `true` to auto-run migrations |
| `JWT_SECRET` | Secret key for JWT signing |
| `GOOGLE_CLIENT_ID` | Google OAuth 2.0 client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth 2.0 client secret |
| `GOOGLE_REDIRECT_URL` | OAuth callback URL |
| `FRONTEND_URL` | Allowed CORS origin |

### 3. Run the server

```bash
go run ./cmd/server/main.go
```

Server starts at `http://localhost:3001`.

---

## Running with Docker Compose

### 1. Start services

```bash
docker-compose up -d
```

This starts:

- `geofencing_db` — PostgreSQL 15 + PostGIS on port `5432`
- `geofencing_backend` — Go API on port `3001`

### 2. View logs

```bash
docker-compose logs -f app
```

### 3. Stop services

```bash
docker-compose down
```

> **Note:** The `app` service depends on the database health check. Migrations run automatically when `MIGRATION=true`.

---

## API Testing Guide

All endpoints are under `http://localhost:3001`.

### Health Check

```bash
curl http://localhost:3001/
```

### Authentication

**Register:**
```bash
curl -X POST http://localhost:3001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:3001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

Save the returned `token` for authenticated requests below:

```bash
TOKEN="<your-jwt-token>"
```

### Geofences

**Create a geofence (polygon):**
```bash
curl -X POST http://localhost:3001/api/v1/geofences/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Warehouse Zone",
    "description": "Main warehouse perimeter",
    "category": "restricted",
    "coordinates": [[77.5946, 12.9716], [77.6046, 12.9716], [77.6046, 12.9816], [77.5946, 12.9816], [77.5946, 12.9716]],
    "created_by": "user-id"
  }'
```

**List geofences:**
```bash
curl http://localhost:3001/api/v1/geofences \
  -H "Authorization: Bearer $TOKEN"
```

**Delete a geofence:**
```bash
curl -X DELETE http://localhost:3001/api/v1/geofences/<geofence-id> \
  -H "Authorization: Bearer $TOKEN"
```

### Vehicles

**Register a vehicle:**
```bash
curl -X POST http://localhost:3001/api/v1/vehicles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "vehicle_number": "KA-01-AB-1234",
    "driver_name": "Rajesh Kumar",
    "vehicle_type": "truck",
    "phone": "+919876543210"
  }'
```

**List vehicles:**
```bash
curl http://localhost:3001/api/v1/vehicles \
  -H "Authorization: Bearer $TOKEN"
```

**Delete a vehicle:**
```bash
curl -X DELETE http://localhost:3001/api/v1/vehicles/<vehicle-id> \
  -H "Authorization: Bearer $TOKEN"
```

### Location

**Update vehicle location:**
```bash
curl -X POST http://localhost:3001/api/v1/vehicles/location \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "vehicle_id": "<vehicle-id>",
    "latitude": 12.976,
    "longitude": 77.599,
    "timestamp": "2025-06-15T10:30:00Z"
  }'
```

**Get vehicle location:**
```bash
curl http://localhost:3001/api/v1/vehicles/location/<vehicle-id> \
  -H "Authorization: Bearer $TOKEN"
```

### Alerts

**Configure an alert:**
```bash
curl -X POST http://localhost:3001/api/v1/alerts/configure \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "geofence_id": "<geofence-id>",
    "vehicle_id": "<vehicle-id>",
    "event_type": "entry"
  }'
```

Event types: `entry`, `exit`, `both`.

**List alerts:**
```bash
curl http://localhost:3001/api/v1/alerts \
  -H "Authorization: Bearer $TOKEN"
```

**Get alert events:**
```bash
curl "http://localhost:3001/api/v1/alert-events?limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

**Delete an alert:**
```bash
curl -X DELETE http://localhost:3001/api/v1/alerts/<alert-id> \
  -H "Authorization: Bearer $TOKEN"
```

### Violations

**Get violation history:**
```bash
curl "http://localhost:3001/api/v1/violations/history?limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

Optional filters: `vehicle_id`, `geofence_id`, `start_date`, `end_date`.

### TrackMe API

**Create a TrackMe entry:**
```bash
curl -X POST http://localhost:3001/api/v1/trackme \
  -H "Content-Type: application/json" \
  -d '{
    "vehicle_number": "KA-01-AB-1234",
    "driver_name": "Rajesh Kumar",
    "vehicle_type": "truck",
    "phone": "+919876543210",
    "status": "active"
  }'
```

**Get TrackMe matches (by vehicle_number or phone):**
```bash
curl "http://localhost:3001/api/v1/trackme/matches?q=KA-01-AB-1234"
```

**Delete a TrackMe entry:**
```bash
curl -X DELETE http://localhost:3001/api/v1/trackme/<id>
```

### WebSocket

**Connect to alerts stream:**
```bash
# Use a WebSocket client (e.g., websocat or browser)
ws://localhost:3001/ws/alerts?token=<jwt-token>
```

**Connect to GPS tracking stream (for Flutter app):**
```bash
ws://localhost:3001/ws/track
```

---

## Architecture Overview

```
┌─────────────────┐     ┌──────────────────────┐     ┌──────────────┐
│   React         │     │   Go Fiber API       │     │  PostgreSQL  │
│   Frontend      │◄───►│   (geofencing_backend)│◄───►│  + PostGIS   │
│   (Vite)        │ WS  │                      │     │              │
└─────────────────┘     │  ┌────────────────┐  │     └──────────────┘
                        │  │  Geofence Svc  │  │
┌─────────────────┐     │  ├────────────────┤  │
│   Flutter       │◄───►│  │  Location Svc  │  │
│   GPS Tracker   │ WS  │  ├────────────────┤  │
└─────────────────┘     │  │  Alert Engine  │  │
                        │  ├────────────────┤  │
                        │  │  Violation Log │  │
                        │  └────────────────┘  │
                        └──────────────────────┘
```

### Layers

- **Handlers** — HTTP request/response (`internal/handlers/`)
- **Routes** — endpoint definitions (`internal/routes/`)
- **Services** — business logic (geofence validation, coordinate math, alert evaluation) (`internal/services/`)
- **Models** — GORM database models (`internal/models/`)
- **DTOs** — request/response structures (`internal/dto/`)

### Real-time Flow

1. Flutter app streams GPS coordinates to `/ws/track`.
2. Backend processes each frame: looks up vehicle, checks against active geofences.
3. On geofence entry/exit, an `AlertEvent` is created and pushed to all connected frontend clients via `/ws/alerts`.
4. Violations are logged in the `violations` table for historical queries.
