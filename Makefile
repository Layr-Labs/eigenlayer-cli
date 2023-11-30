.PHONY: help build test

include .env

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Compile the binary
	@mkdir -p bin
	@go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

mocks: ## generates mocks
	go install go.uber.org/mock/mockgen@v0.3.0
	go generate ./...

test: ## runs all tests
	go test ./... -covermode=atomic