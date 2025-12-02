GOCACHE ?= $(CURDIR)/.gocache

run:
	GOCACHE=$(GOCACHE) go run cmd/api/main.go

sqlc:
	sqlc generate

# migrate-up:
# 	# Add migration command here if using a tool like migrate
# 	echo "Migrations disabled by user request"

test:
	go test ./...

docker-run:
	# Levanta API y Postgres con docker-compose
	docker compose --env-file .env.docker up -d --build api

docker-down:
	# Detiene y borra contenedores/volúmenes named
	docker compose down
