APP_CMD=./cmd/server
MIGRATE_CMD=./cmd/migration
SEED_CMD=./cmd/seed
RENAME_CMD=./cmd/rename
GEN_CMD=./cmd/generate
SWAGGER_OUT=internal/server/swagger/docs
SWAGGER_ENTRY=$(APP_CMD)/main.go

.PHONY: run migrate-up migrate-down migrate-seed build test rename generate swagger

run:
	@echo "🚀 Starting server..."
	@set -a; [ -f .env ] && . ./.env; set +a; go run $(APP_CMD)

migrate-up:
	@echo "⬆️  Applying database migrations..."
	@go run $(MIGRATE_CMD) up

migrate-down:
	@echo "⬇️  Rolling back database migrations..."
	@go run $(MIGRATE_CMD) down

migrate-seed:
	@echo "🌱 Seeding database..."
	@go run $(SEED_CMD)

build:
	@echo "🏗️  Building server binary..."
	@mkdir -p bin
	@go build -o bin/server $(APP_CMD)
test:
	@echo "🧪 Running tests..."
	@go test ./...

rename:
	@if [ -z "$(old)" ] || [ -z "$(new)" ]; then \
		echo "Usage: make rename old=<current module> new=<new module>"; \
		exit 1; \
	fi
	@echo "✏️  Renaming module from $(old) to $(new)..."
	@go run $(RENAME_CMD) -old $(old) -new $(new)

generate:
	@if [ -z "$(word 2,$(MAKECMDGOALS))" ] || [ -z "$(word 3,$(MAKECMDGOALS))" ]; then \
		echo "Usage: make generate modules <name>"; \
		exit 1; \
	fi
	@echo "🛠️  Generating $(word 2,$(MAKECMDGOALS)) $(word 3,$(MAKECMDGOALS))..."
	@go run $(GEN_CMD) $(word 2,$(MAKECMDGOALS)) $(word 3,$(MAKECMDGOALS))

swagger:
	@echo "📜 Updating Swagger docs..."
	@go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g $(SWAGGER_ENTRY) -o $(SWAGGER_OUT) --parseInternal

%::
	@:
