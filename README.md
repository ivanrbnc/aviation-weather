# ✈️ Aviation Weather API

> Merge airport data with real-time weather in one API

Built with Go, combining [Aviation API](https://www.aviationapi.com/) and [Weather API](https://www.weatherapi.com/)

## 🚀 Quick Start

```bash
# Start the server
docker-compose up --build

# Initialize database
docker-compose exec app go run cmd/migration/main.go --fill
```

API available at `http://localhost:8080`

## 📡 Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `localhost:8080/airports` | List all airports |
| `GET` | `localhost:8080/airport/{faa}` | Get airport from database |
| `POST` | `localhost:8080/airport` | Create airport |
| `PUT` | `localhost:8080/airport/{faa}` | Update airport |
| `DELETE` | `localhost:8080/airport/{faa}` | Delete airport |
| `POST` | `localhost:8080/sync/{faa}` | Sync single airport |
| `POST` | `localhost:8080/sync` | Sync all airport |

## 🧪 Try It Out
Import `Aviation Weather.postman_collection.json` into Postman to test all endpoints!

## 🔧 Config

Create `.env`:
```env
# DB
DB_HOST=host.docker.internal
DB_PORT=5430
DB_NAME=aviation_weather
DB_USER=postgres
DB_PASSWORD=postgres

# APIs
WEATHER_API_KEY=YOUR_WEATHER_API_KEY

# App
APP_PORT=8080
```
---

Made with ❤️ using Go