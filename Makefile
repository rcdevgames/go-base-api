APP_CMD=./cmd/server
MIGRATE_CMD=./cmd/migration
SEED_CMD=./cmd/seed
RENAME_CMD=./cmd/rename
GEN_CMD=./cmd/generate

.PHONY: run migrate-up migrate-down migrate-seed build test rename generate

run:
	go run $(APP_CMD)

migrate-up:
	go run $(MIGRATE_CMD) up

migrate-down:
	go run $(MIGRATE_CMD) down

migrate-seed:
	go run $(SEED_CMD)

build:
	@mkdir -p bin
	go build -o bin/server $(APP_CMD)

test:
	go test ./...

rename:
	@if [ -z "$(old)" ] || [ -z "$(new)" ]; then \
		echo "Usage: make rename old=<current module> new=<new module>"; \
		exit 1; \
	fi
	go run $(RENAME_CMD) -old $(old) -new $(new)

generate:
	@if [ -z "$(word 2,$(MAKECMDGOALS))" ] || [ -z "$(word 3,$(MAKECMDGOALS))" ]; then \
		echo "Usage: make generate modules <name>"; \
		exit 1; \
	fi
	go run $(GEN_CMD) $(word 2,$(MAKECMDGOALS)) $(word 3,$(MAKECMDGOALS))

%::
	@:
