SHELL := /bin/sh

COMPOSE ?= docker compose
POSTGRES_SERVICE ?= postgres
POSTGRES_USER ?= learning_english
POSTGRES_DB ?= learning_english
MIGRATE_SERVICE ?= migrate
MIGRATIONS_DIR ?= backend/migrations
STEPS ?= 1

.PHONY: help compose-config stack-up stack-down stack-logs stack-ps db-up db-down db-logs db-ps db-health db-shell migrate-create migrate-up migrate-down migrate-version migrate-force

help:
	@printf '%s\n' \
		'Available targets:' \
		'  compose-config  Validate the Compose file' \
		'  stack-up        Start the local frontend, backend, and PostgreSQL services' \
		'  stack-down      Stop the local Compose stack' \
		'  stack-logs      Follow frontend, backend, and PostgreSQL logs' \
		'  stack-ps        Show the full Compose service status' \
		'  db-up           Start the local PostgreSQL service' \
		'  db-down         Stop the local Compose stack' \
		'  db-logs         Follow PostgreSQL logs' \
		'  db-ps           Show Compose service status' \
		'  db-health       Check PostgreSQL readiness from inside the container' \
		'  db-shell        Open a psql shell against the local database' \
		'  migrate-create  Create a new sequential SQL migration pair (NAME=...)' \
		'  migrate-up      Apply all pending migrations' \
		'  migrate-down    Roll back migration steps (STEPS=1 by default)' \
		'  migrate-version Show the current migration version' \
		'  migrate-force   Force the migration version after dirty-state recovery (VERSION=...)'

compose-config:
	$(COMPOSE) config

stack-up:
	$(COMPOSE) up -d postgres backend frontend

stack-down:
	$(COMPOSE) down

stack-logs:
	$(COMPOSE) logs -f postgres backend frontend

stack-ps:
	$(COMPOSE) ps

db-up:
	$(COMPOSE) up -d $(POSTGRES_SERVICE)

db-down:
	$(COMPOSE) down

db-logs:
	$(COMPOSE) logs -f $(POSTGRES_SERVICE)

db-ps:
	$(COMPOSE) ps

db-health:
	$(COMPOSE) exec $(POSTGRES_SERVICE) pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB)

db-shell:
	$(COMPOSE) exec $(POSTGRES_SERVICE) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

migrate-create:
	@test -n "$(NAME)" || { printf '%s\n' 'NAME is required, e.g. make migrate-create NAME=create_users'; exit 1; }
	@mkdir -p $(MIGRATIONS_DIR)
	$(COMPOSE) run --rm --no-deps --entrypoint migrate $(MIGRATE_SERVICE) create -ext sql -dir /migrations -seq $(NAME)

migrate-up: db-up
	@mkdir -p $(MIGRATIONS_DIR)
	$(COMPOSE) run --rm $(MIGRATE_SERVICE) up

migrate-down: db-up
	$(COMPOSE) run --rm $(MIGRATE_SERVICE) down $(STEPS)

migrate-version: db-up
	$(COMPOSE) run --rm $(MIGRATE_SERVICE) version

migrate-force: db-up
	@test -n "$(VERSION)" || { printf '%s\n' 'VERSION is required, e.g. make migrate-force VERSION=1'; exit 1; }
	$(COMPOSE) run --rm $(MIGRATE_SERVICE) force $(VERSION)
