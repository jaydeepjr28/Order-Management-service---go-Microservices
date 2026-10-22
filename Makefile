.PHONY: all run deps up-infra up frontend down

all: deps run frontend

# Infrastructure dependencies (Kafka, Consul, Jaeger, Postgres, Redis)
deps up-infra:
	docker compose -f docker-compose.dev.yml up -d kafka kafka-ui consul jaeger gateway-postgres orders-postgres payments-postgres stock-postgres redis

# Full stack in Docker containers
up:
	docker compose -f docker-compose.dev.yml up -d

run:
	@echo "Starting backend microservices..."
	powershell -ExecutionPolicy Bypass -File ./scripts/start-services.ps1

frontend:
	@echo "Starting frontend development server..."
	powershell -ExecutionPolicy Bypass -Command "Start-Process powershell -ArgumentList '-NoExit', '-Command', 'cd frontend; npm run dev'"

down:
	docker compose -f docker-compose.dev.yml down

stop:
	@echo "Stopping backend microservices..."
	powershell -ExecutionPolicy Bypass -File ./scripts/stop-services.ps1
