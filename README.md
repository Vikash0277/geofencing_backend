# Geofencing Backend

Go-based REST API + WebSocket backend for geofencing, vehicle tracking, and alert management.

## Tech Stack

- **Framework:** Go (Fiber v2)
- **Database:** PostgreSQL + PostGIS
- **ORM:** GORM
- **Auth:** JWT + Google OAuth 2.0
- **Maps:** go-geom for geospatial operations

## Features

- Geofence CRUD (polygon-based boundaries)
- Vehicle registration and management
- Real-time location updates with geofence entry/exit detection
- Configurable alerts (per geofence, per vehicle)
- Violation history logging
- WebSocket streams for live alerts and GPS tracking
- TrackMe API for Flutter app integration
- Google OAuth single sign-on
