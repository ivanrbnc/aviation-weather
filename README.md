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

### By Docker & Kubernetes
```bash
# Install NGINX Ingress Controller (Once)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.14.1/deploy/static/provider/cloud/deploy.yaml

# Wait for the ingress-nginx-controller running
kubectl get pods -n ingress-nginx

# Activate Postgresql
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

### By Docker & Kubernetes via Jenkins
```bash
# Install NGINX Ingress Controller (Once)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml

# Wait for the ingress-nginx-controller running
kubectl get pods -n ingress-nginx

# Activate Postgresql
docker-compose up --build -d postgres

# Create a folder outside aviation-weather. Copy the docker-compose.yml from jenkins-setup.
# Make sure to update the docker-compose.yml

# Up the Jenkins
docker-compose up -d

# Get initial admin password to login the localhost:8090
docker exec jenkins cat /var/jenkins_home/secrets/initialAdminPassword

# Install docker cli to your jenkins
docker exec -u root jenkins bash -c "apt-get update && apt-get install -y docker.io"
docker exec -u root jenkins chmod 666 /var/run/docker.sock

# Install kubectl to your jenkins
docker exec -u root jenkins bash -c "apt-get update && apt-get install -y docker.io curl && curl -LO 'https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl' && chmod +x kubectl && mv kubectl /usr/local/bin/kubectl && chmod 666 /var/run/docker.sock"

# Restart jenkins
docker restart jenkins

# Open http://localhost:8090/manage/pluginManager/available for manage plugin
# Make sure "Kubernetes CLI Plugin", "Docker pipeline", "Pipeline: stage view" was installed

# Open http://localhost:8090/manage/credentials/ for manage credentials
# Kind: Secret file
# File: Kube's config file. Example: C:\Users\vyanry\.kube\config
# ID: kubeconfig
# Description: Kubernetes Config

# Open http://localhost:8090/view/all/newJob for deployments
# Item name: aviation-weather-deploy
# Item type: pipeline
# Configure > Pipeline > Definition: Pipeline script. Copy jenkins-setup/deployment/Jenkinsfile content here.

# Click `Build Now`

# Create docker image
docker build -t aviation-weather-service:v1 .

# Open http://localhost:8090/view/all/newJob for testing
# Item name: aviation-weather-test
# Item type: pipeline
# Configure > Pipeline > Definition: Pipeline script. Copy jenkins-setup/testing/Jenkinsfile content here.

# To delete all kubernetes enabled as aviation-weather
    kubectl delete all,ingress,cronjob,pvc,configmap,secret --all -n aviation-weather
```

### Report via Docker & k6
```bash
# Start the server
docker-compose up --build

# Initialize database
docker-compose exec app go run cmd/migration/main.go --fill

# Pull image `k6` in Docker Hub
docker pull grafana/k6

# Run k6 via `k6` image
# Update D:\vyanry\aviation-weather to your aviation-weather path
docker run --rm -v "D:\vyanry\aviation-weather:/scripts" `
  -e K6_WEB_DASHBOARD=true `
  -e K6_WEB_DASHBOARD_EXPORT=/scripts/test/report.html `
  -p 5665:5665 `
  grafana/k6 run /scripts/test/aviation_weather_test.js

# Open http://localhost:5665/ while it's running
```

- API available at http://localhost:8080 for Docker
- API available at http://localhost for Docker + Kubernetes or Docker + Jenkins
- (optional) Jenkins available at http://localhost:8090
- (optional) Report k6 available at http://localhost:5665/

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
