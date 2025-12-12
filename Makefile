SHELL := /bin/bash
GOCACHE ?= $(CURDIR)/.gocache

# Env files (override from CLI: ENVFILE_DEV=.env.local make run-dev)
ENVFILE_DEV ?= .env.dev
ENVFILE_PROD ?= .env

# Helper to load env vars from a file (ignores comments/blank lines and strips CRLF)
define load_env
	env $$(sed -e '/^\s*#/d' -e '/^\s*$$/d' -e 's/\r$$//' $(1) | xargs)
endef

run:
	GOCACHE=$(GOCACHE) go run cmd/api/main.go

run-dev:
	$(call load_env,$(ENVFILE_DEV)) APP_ENV=development GOCACHE=$(GOCACHE) go run cmd/api/main.go

run-prod:
	$(call load_env,$(ENVFILE_PROD)) APP_ENV=production GOCACHE=$(GOCACHE) go run cmd/api/main.go

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
