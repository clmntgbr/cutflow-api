COMPOSE_DEV := docker compose -f compose.dev.yaml

# ============================================
# Development (compose.dev.yaml)
# ============================================

dev:
	$(COMPOSE_DEV) up -d

restart:
	$(COMPOSE_DEV) restart worker api extraction silence transcript analysis

build:
	$(COMPOSE_DEV) up -d --build

dev-down:
	$(COMPOSE_DEV) down

dev-logs:
	$(COMPOSE_DEV) logs -f api worker extraction silence transcript analysis

api-logs:
	$(COMPOSE_DEV) logs -f api

worker-logs:
	$(COMPOSE_DEV) logs -f worker

extraction-logs:
	$(COMPOSE_DEV) logs -f extraction

dev-restart:
	$(COMPOSE_DEV) restart api worker extraction silence transcript analysis

lint:
	$(COMPOSE_DEV) exec api golangci-lint run --fix

# ============================================
# Tests (handler HTTP, via api container)
# ============================================

tests:
	$(COMPOSE_DEV) exec api go test ./internal/interfaces/http/handler/test/... -v -count=1

coverage:
	$(COMPOSE_DEV) exec api go test -coverprofile=coverage.out ./internal/interfaces/http/handler/test/... -coverpkg=./internal/interfaces/http/handler/...

coverage-html: coverage
	$(COMPOSE_DEV) exec api go tool cover -html=coverage.out -o coverage.html

# ============================================
# CLI Commands (via Docker)
# ============================================

cli-build:
	@$(COMPOSE_DEV) exec api go build -o bin/cli ./cmd/cli

migrate: cli-build
	@echo "Applying SQL migrations..."
	@$(COMPOSE_DEV) exec api ./bin/cli migrate up
	@echo "Checking model/DB schema..."
	@$(COMPOSE_DEV) exec api ./bin/cli migrate check

migrate-status: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli migrate status

migrate-down: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli migrate down

migrate-check: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli migrate check

purge: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli purge --yes

purge-storage: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli purge-storage --yes

build-transcript-fixture: cli-build
	@$(COMPOSE_DEV) exec api ./bin/cli build-transcript-fixture \
		--srt=fixtures/subtitles.srt \
		--out=fixtures/transcript.json \
		--language=fr

export-transcript: cli-build
	@test -n "$(MEDIA_FILE_ID)" || (echo "MEDIA_FILE_ID required"; exit 1)
	@$(COMPOSE_DEV) exec api ./bin/cli export-transcript \
		--media-file-id=$(MEDIA_FILE_ID) \
		--out=fixtures/transcript.json

shell:
	$(COMPOSE_DEV) exec api sh
