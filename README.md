# ✈️ Aviation Weather API

> Merge airport data with real-time weather in one API

Built with Go, combining [Aviation API](https://www.aviationapi.com/) and [Weather API](https://www.weatherapi.com/)

## 🚀 Quick Start

### By Docker
```bash
# Start the server
docker-compose up --build

# Initialize database
docker-compose exec app go run cmd/migration/main.go --fill
```

### By Kubernetes & Docker
```bash
# Install NGINX Ingress Controller (Once)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml

# Wait for the ingress-nginx-controller running
kubectl get pods -n ingress-nginx

# Activate Postgresql & App
docker-compose up --build -d postgres

# Create docker image
docker build -t aviation-weather-service:v1 .

# Initialize kubernetes pods by configuration.
kubectl apply -k k8s/

# Wait for the db-migrate-and-seed completed
kubectl get pods -n aviation-weather

# Monitoring the server
kubectl logs -f deployment/aviation-weather-deployment -c server -n aviation-weather

# To delete all kubernetes enabled as aviation-weather
kubectl delete all,ingress,cronjob,pvc,configmap,secret --all -n aviation-weather
```

- API available at http://localhost:8080 for Docker
- API available at http://localhost for Docker + Kubernetes

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

or

Update `k8s/secret.yaml` and `k8s/configmap.yaml`

---

Made with Go, Docker, Kubernetes, and Postgresql
