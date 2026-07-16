# Geofencing Backend

## Overview
A Go‑based REST API + WebSocket service that provides geofencing, vehicle tracking, and alert management capabilities. It powers the companion Flutter "TrackMe" app and offers a robust set of endpoints for managing geofences, vehicles, and real‑time location updates.

## Table of Contents
- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Service](#running-the-service)
- [API Reference](#api-reference)
- [WebSocket Streams](#websocket-streams)
- [Testing](#testing)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Tech Stack
- **Language / Framework:** Go (Fiber v2)
- **Database:** PostgreSQL with PostGIS for spatial queries
- **ORM:** GORM
- **Authentication:** JWT & Google OAuth 2.0
- **Geospatial Library:** go‑geom
- **Containerisation:** Docker & Docker‑Compose

## Features
- CRUD operations for polygon‑based geofences
- Vehicle registration & management
- Real‑time location ingest with geofence entry/exit detection
- Configurable alerts per geofence / vehicle
- Violation history persistence
- WebSocket streams for live alerts & GPS tracking
- Seamless integration with the Flutter **TrackMe** client
- Google OAuth single‑sign‑on for admin UI

## Prerequisites
- Go 1.22 or newer
- Docker & Docker‑Compose (for containerised development)
- PostgreSQL 13+ with the PostGIS extension
- (Optional) Make for convenience scripts

## Installation
```bash
# Clone the repository
git clone https://github.com/your-org/geofencing_backend.git
cd geofencing_backend

# Build the Go binary
make build   # or: go build -o bin/server ./cmd/server

# Or use Docker
docker compose build
```

## Configuration
Create a `.env` file in the project root (copy from `.env.example`):
```dotenv
# Server
PORT=8080
JWT_SECRET=your-secret-key
GOOGLE_OAUTH_CLIENT_ID=xxxx.apps.googleusercontent.com
GOOGLE_OAUTH_CLIENT_SECRET=xxxxxxxxxx

# Database
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=geofence
POSTGRES_USER=geofence_user
POSTGRES_PASSWORD=securepassword
POSTGIS_ENABLED=true
```
The service reads these variables at startup.

## Running the Service
### Locally (without Docker)
```bash
go run ./cmd/server
```
### With Docker Compose
```bash
docker compose up -d
```
The API will be available at `http://localhost:8080`.

## API Reference
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/geofences` | Create a new geofence (polygon) |
| `GET`  | `/api/v1/geofences` | List all geofences |
| `GET`  | `/api/v1/geofences/:id` | Retrieve a single geofence |
| `PUT`  | `/api/v1/geofences/:id` | Update geofence geometry or metadata |
| `DELETE`| `/api/v1/geofences/:id` | Delete a geofence |
| `POST` | `/api/v1/vehicles` | Register a vehicle |
| `GET`  | `/api/v1/vehicles/:id/location` | Get latest location for a vehicle |
| `POST` | `/api/v1/vehicles/:id/location` | Submit a new location (triggers geofence checks) |
| `GET`  | `/api/v1/alerts` | List recent alerts |

Full Swagger/OpenAPI documentation is served at `/swagger/index.html` when the server is running.

## WebSocket Streams
- **Endpoint:** `ws://localhost:8080/ws/alerts`
- Subscribe to real‑time alert events (entry/exit, violations).
- Message format (JSON): `{ "type": "alert", "vehicle_id": "...", "geofence_id": "...", "event": "enter|exit", "timestamp": "..." }`

## Testing
```bash
# Unit tests
go test ./... -v

# Integration tests (requires Docker services)
make test-integration
```
Coverage reports are generated in `coverage/`.

## Development
- Use `make fmt` and `make lint` to keep code style consistent.
- Run `make dev` to start a live‑reload development server with Docker.
- Database migrations are managed via `golang-migrate`. Run `make migrate-up` to apply.

## Contributing
Contributions are welcome! Please follow these steps:
1. Fork the repository.
2. Create a feature branch (`git checkout -b feat/your-feature`).
3. Write tests for your changes.
4. Ensure `go vet`, `golint`, and `make fmt` pass.
5. Open a Pull Request with a clear description of the change.

## License
This project is licensed under the **MIT License** – see the [LICENSE](LICENSE) file for details.
