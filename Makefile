ifneq (,$(wildcard .env))
	include .env
	export $(shell sed 's/=.*//' .env)
endif

COVERAGE_FILE ?= coverage.out

.PHONY: build
build: build_bot build_scrapper

.PHONY: build_bot
build_bot:
	@echo "Выполняется go build для таргета bot"
	@mkdir -p .bin
	@go build -o ./bin/bot ./cmd/bot

.PHONY: build_scrapper
build_scrapper:
	@echo "Выполняется go build для таргета scrapper"
	@mkdir -p .bin
	@go build -o ./bin/scrapper ./cmd/scrapper

.PHONY: run_bot
run_bot:
	@go run ./cmd/bot/main.go

.PHONY: run_scrapper
run_scrapper:
	@go run ./cmd/scrapper/main.go

## test: run all tests
.PHONY: test
test:
	@go test -coverpkg='github.com/es-debug/backend-academy-2024-go-template/...' --race -count=1 -coverprofile='$(COVERAGE_FILE)' ./...
	@go tool cover -func='$(COVERAGE_FILE)' | grep ^total | tr -s '\t'

.PHONY: lint
lint: lint-golang lint-proto

.PHONY: lint-golang
lint-golang:
	@if ! command -v 'golangci-lint' &> /dev/null; then \
  		echo "Please install golangci-lint!"; exit 1; \
  	fi;
	@golangci-lint -v run --fix ./...

.PHONY: lint-proto
lint-proto:
	@if ! command -v 'easyp' &> /dev/null; then \
  		echo "Please install easyp!"; exit 1; \
	fi;
	@easyp lint

.PHONY: generate
generate: generate_proto generate_openapi

.PHONY: generate_proto
generate_proto:
	@if ! command -v 'easyp' &> /dev/null; then \
		echo "Please install easyp!"; exit 1; \
	fi;
	@easyp generate

.PHONY: generate_openapi
generate_openapi:
	@if ! command -v 'oapi-codegen' &> /dev/null; then \
		echo "Please install oapi-codegen!"; exit 1; \
	fi;
	@mkdir -p internal/api/openapi/v1
	@oapi-codegen -package v1 \
		-generate server,types \
		api/openapi/v1/service.yaml > internal/api/openapi/v1/service.gen.go

.PHONY: clean
clean:
	@rm -rf./bin

.PHONY: migrate_bot
migrate_bot:
	@go run ./cmd/bot_migration/main.go --command up

.PHONY: migrate_scrapper
migrate_scrapper:
	@go run ./cmd/scrapper_migration/main.go --command up

.PHONY: run
run:
	@docker-compose up -d --build --force-recreate --remove-orphans

.PHONY: local
local:
	@docker-compose -f docker-compose.local.yml up -d --build --force-recreate --remove-orphans

.PHONY: cover
cover:
	@go tool cover -html='$(COVERAGE_FILE)'