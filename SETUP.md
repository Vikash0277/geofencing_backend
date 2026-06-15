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

All endpoints are under `http://localhost:3001`. Authenticated endpoints require a JWT token in the `Authorization` header.

### Complete Endpoint Reference

| # | Method | Path | Auth | Description |
|---|--------|------|------|-------------|
| 1 | `GET` | `/` | No | Health check |
| 2 | `POST` | `/api/v1/auth/register` | No | Register a new user |
| 3 | `POST` | `/api/v1/auth/login` | No | Login with email/password |
| 4 | `GET` | `/api/v1/auth/google` | No | Initiate Google OAuth login |
| 5 | `GET` | `/api/v1/auth/google/callback` | No | Google OAuth callback handler |
| 6 | `POST` | `/api/v1/geofences/` | Yes | Create a polygon geofence |
| 7 | `GET` | `/api/v1/geofences` | Yes | List geofences (optional `?category=`) |
| 8 | `DELETE` | `/api/v1/geofences/:id` | Yes | Delete a geofence |
| 9 | `POST` | `/api/v1/vehicles` | Yes | Register a vehicle |
| 10 | `GET` | `/api/v1/vehicles` | Yes | List all vehicles |
| 11 | `DELETE` | `/api/v1/vehicles/:id` | Yes | Delete a vehicle |
| 12 | `POST` | `/api/v1/vehicles/location` | Yes | Update vehicle GPS location |
| 13 | `GET` | `/api/v1/vehicles/location/:vehicle_id` | Yes | Get latest location of a vehicle |
| 14 | `POST` | `/api/v1/alerts/configure` | Yes | Configure an alert rule |
| 15 | `GET` | `/api/v1/alerts` | Yes | List alert configs (optional `?geofence_id=`, `?vehicle_id=`) |
| 16 | `DELETE` | `/api/v1/alerts/:id` | Yes | Delete an alert config |
| 17 | `GET` | `/api/v1/alert-events` | Yes | Get alert events (optional `?limit=`, default 50, max 500) |
| 18 | `GET` | `/api/v1/violations/history` | Yes | Get violation history with filters |
| 19 | `POST` | `/api/v1/trackme` | No | Create a TrackMe entry |
| 20 | `DELETE` | `/api/v1/trackme/:id` | No | Delete a TrackMe entry |
| 21 | `GET` | `/api/v1/trackme/matches` | No | Get TrackMe entries matched to vehicles |
| 22 | `GET` | `/ws/alerts` | Token* | WebSocket stream for real-time alerts |
| 23 | `GET` | `/ws/track` | No | WebSocket endpoint for Flutter GPS tracking |

> `*` — `/ws/alerts` requires a `?token=` query parameter instead of the Authorization header.

---

### 1. Health Check

```bash
curl http://localhost:3001/
```

**Response:**
```json
{
  "message": "Welcome to Geofencing Project",
  "status": "running"
}
```

---

### 2–3. Authentication (Register / Login)

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

**Response (both):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid-here",
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user",
    "provider": "local"
  }
}
```

Save the token for authenticated requests:

```bash
TOKEN="eyJhbGciOiJIUzI1NiIs..."
```

---

### 4–5. Google OAuth

**Initiate Google login (opens Google consent screen):**
```bash
# Open in browser — redirects to Google
curl http://localhost:3001/api/v1/auth/google
```

After successful consent, Google redirects to `/api/v1/auth/google/callback`, which exchanges the code for user info, creates/links the user, and redirects the browser to `FRONTEND_URL#token=<jwt>&user=<json>`.

---

### 6–8. Geofences

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

- `coordinates` — array of `[longitude, latitude]` pairs forming a closed polygon (first and last must match).
- `category` — any label (e.g., `restricted`, `warehouse`, `school-zone`).

**List geofences (with optional category filter):**
```bash
# All geofences
curl http://localhost:3001/api/v1/geofences \
  -H "Authorization: Bearer $TOKEN"

# Filter by category
curl "http://localhost:3001/api/v1/geofences?category=restricted" \
  -H "Authorization: Bearer $TOKEN"
```

**Delete a geofence (cascades to related alerts, events, violations):**
```bash
curl -X DELETE http://localhost:3001/api/v1/geofences/<geofence-id> \
  -H "Authorization: Bearer $TOKEN"
```

---

### 9–11. Vehicles

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

**List vehicles (only those matched to TrackMe entries):**
```bash
curl http://localhost:3001/api/v1/vehicles \
  -H "Authorization: Bearer $TOKEN"
```

**Delete a vehicle (cascades to locations, alerts, events, violations):**
```bash
curl -X DELETE http://localhost:3001/api/v1/vehicles/<vehicle-id> \
  -H "Authorization: Bearer $TOKEN"
```

---

### 12–13. Location

**Update vehicle location (triggers geofence evaluation + alert generation):**
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

**Response includes:** geofence status, entry/exit events, and triggered alerts.

**Get latest vehicle location:**
```bash
curl http://localhost:3001/api/v1/vehicles/location/<vehicle-id> \
  -H "Authorization: Bearer $TOKEN"
```

**Response includes:** current lat/lng, timestamp, and list of geofences the vehicle is currently inside.

---

### 14–17. Alerts

**Configure an alert rule:**
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

- `event_type` — one of `entry`, `exit`, `both`.
- `vehicle_id` — optional; omit to alert on any vehicle entering/exiting the geofence.

**List alert configurations (with optional filters):**
```bash
# All alert configs
curl http://localhost:3001/api/v1/alerts \
  -H "Authorization: Bearer $TOKEN"

# Filter by geofence
curl "http://localhost:3001/api/v1/alerts?geofence_id=<geofence-id>" \
  -H "Authorization: Bearer $TOKEN"

# Filter by vehicle
curl "http://localhost:3001/api/v1/alerts?vehicle_id=<vehicle-id>" \
  -H "Authorization: Bearer $TOKEN"
```

**Get alert events (recent triggered alerts):**
```bash
# Last 50 events (default)
curl http://localhost:3001/api/v1/alert-events \
  -H "Authorization: Bearer $TOKEN"

# Custom limit (max 500)
curl "http://localhost:3001/api/v1/alert-events?limit=100" \
  -H "Authorization: Bearer $TOKEN"
```

**Delete an alert config:**
```bash
curl -X DELETE http://localhost:3001/api/v1/alerts/<alert-id> \
  -H "Authorization: Bearer $TOKEN"
```

---

### 18. Violations

**Get violation history with filters:**
```bash
# Last 50 violations
curl http://localhost:3001/api/v1/violations/history \
  -H "Authorization: Bearer $TOKEN"

# Custom limit (max 500)
curl "http://localhost:3001/api/v1/violations/history?limit=20" \
  -H "Authorization: Bearer $TOKEN"

# Filter by vehicle
curl "http://localhost:3001/api/v1/violations/history?vehicle_id=<vehicle-id>" \
  -H "Authorization: Bearer $TOKEN"

# Filter by geofence
curl "http://localhost:3001/api/v1/violations/history?geofence_id=<geofence-id>" \
  -H "Authorization: Bearer $TOKEN"

# Filter by date range
curl "http://localhost:3001/api/v1/violations/history?start_date=2025-01-01&end_date=2025-06-30" \
  -H "Authorization: Bearer $TOKEN"

# Combined filters
curl "http://localhost:3001/api/v1/violations/history?vehicle_id=<id>&geofence_id=<id>&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Available query parameters:**

| Param | Type | Description |
|---|---|---|
| `vehicle_id` | string | Filter by vehicle UUID |
| `geofence_id` | string | Filter by geofence UUID |
| `start_date` | string | Start of date range (RFC3339 or date string) |
| `end_date` | string | End of date range (RFC3339 or date string) |
| `limit` | int | Max results (1–500, default 50) |

---

### 19–21. TrackMe API (Flutter App Integration)

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

**Get TrackMe matches (rows joined to vehicles on vehicle_number OR phone):**
```bash
curl http://localhost:3001/api/v1/trackme/matches
```

**Response:**
```json
{
  "matches": [
    {
      "trackme_id": "uuid",
      "vehicle_number": "KA-01-AB-1234",
      "driver_name": "Rajesh Kumar",
      "vehicle_type": "truck",
      "phone": "+919876543210",
      "trackme_status": "active",
      "vehicle_id": "uuid-for-ws-tracking"
    }
  ],
  "count": 1,
  "time_ns": 12345
}
```

> `vehicle_id` is the UUID the Flutter app uses to connect to the WebSocket tracking stream.

**Delete a TrackMe entry:**
```bash
curl -X DELETE http://localhost:3001/api/v1/trackme/<trackme-id>
```

---

### 22–23. WebSocket Endpoints

**Connect to the real-time alerts stream (frontend uses this):**
```bash
# Use a WebSocket client (e.g., websocat, wscat, or browser DevTools)
ws://localhost:3001/ws/alerts?token=<jwt-token>
```

The server pushes JSON frames whenever a geofence entry/exit event is triggered:

```json
{
  "type": "geofence_alert",
  "alert_id": "uuid",
  "alert_config_id": "uuid",
  "geofence_id": "uuid",
  "geofence_name": "Warehouse Zone",
  "vehicle_id": "uuid",
  "vehicle_number": "KA-01-AB-1234",
  "event_type": "entry",
  "message": "Vehicle KA-01-AB-1234 entered geofence Warehouse Zone",
  "latitude": 12.976,
  "longitude": 77.599,
  "timestamp": "2025-06-15T10:30:00Z"
}
```

**Connect to the GPS tracking stream (Flutter app sends GPS here):**
```bash
ws://localhost:3001/ws/track
```

The Flutter app sends JSON frames:

```json
{
  "vehicle_id": "uuid",
  "latitude": 12.976,
  "longitude": 77.599,
  "timestamp": "2025-06-15T10:30:00Z"
}
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
