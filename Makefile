.PHONY: all run deps

all: deps run

deps:
	docker-compose -f docker-compose.dev.yml up -d rabbitmq consul jaeger mongo

run:
	@echo "Starting services..."
	start cmd /c "cd gateway && go run main.go"
	start cmd /c "cd orders && go run main.go"
	start cmd /c "cd stock && go run main.go"
	start cmd /c "cd payments && go run main.go"
	start cmd /c "cd kitchen && go run main.go"
