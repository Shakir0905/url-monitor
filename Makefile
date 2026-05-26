.PHONY: help up down build logs ps migrate-up migrate-down clean

help:
	@echo "Available targets:"
	@echo "  up           - Start all services"
	@echo "  down         - Stop all services"
	@echo "  build        - Rebuild all service images"
	@echo "  logs         - Tail logs from all services"
	@echo "  ps           - Show service status"
	@echo "  migrate-up   - Apply database migrations"
	@echo "  migrate-down - Rollback last migration"
	@echo "  clean        - Remove all containers and volumes"

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

ps:
	docker compose ps

migrate-up:
	migrate -path migrations -database "postgres://app:app_password@localhost:5432/url_monitor?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://app:app_password@localhost:5432/url_monitor?sslmode=disable" down 1

clean:
	docker compose down -v

.PHONY: install-hooks lint test fmt

install-hooks:
	cp scripts/hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

lint:
	golangci-lint run ./...

test:
	go test -v ./...

fmt:
	gofmt -w .
	goimports -w .

.PHONY: hooks-off hooks-on hooks-status

hooks-off:
	@if [ -f .git/hooks/pre-commit ]; then \
		mv .git/hooks/pre-commit .git/hooks/pre-commit.bak; \
		echo "✓ Pre-commit hook disabled"; \
	else \
		echo "Hook already disabled or not installed"; \
	fi

hooks-on:
	@if [ -f .git/hooks/pre-commit.bak ]; then \
		mv .git/hooks/pre-commit.bak .git/hooks/pre-commit; \
		echo "✓ Pre-commit hook enabled"; \
	else \
		echo "No .bak file found - run 'make install-hooks' instead"; \
	fi

hooks-status:
	@if [ -f .git/hooks/pre-commit ]; then \
		echo "Pre-commit hook: ENABLED"; \
	elif [ -f .git/hooks/pre-commit.bak ]; then \
		echo "Pre-commit hook: DISABLED (.bak exists)"; \
	else \
		echo "Pre-commit hook: NOT INSTALLED"; \
	fi
