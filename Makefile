#!make

SHELL := /bin/sh
GOPATH := $(shell go env GOPATH)
GITROOT := $(shell git rev-parse --show-toplevel)

include $(GITROOT)/.env
export

.PHONY: start-dev
start-dev: ## Starts up a mongodb instance
	docker-compose --env-file=$(GITROOT)/.env up -d

.PHONY: stop
stop: ## Stops
	docker-compose down

.PHONY: swag
swag: ## Builds Swagger Spec Files
	swag init --dir $(GITROOT)/ \
		--output ./internal/docs \
		--generalInfo /cmd/api/main.go \
		--markdownFiles /internal \

.PHONY: lint
lint: ## Runs Linter
	golangci-lint run --timeout 15m
