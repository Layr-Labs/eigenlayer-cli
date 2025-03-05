.PHONY: help build tests mocks fmt format-lines lint install build-linux-amd64 build-linux-arm64 build-linux

include .env

GO_LINES_IGNORED_DIRS=
GO_PACKAGES=./pkg/... ./cmd/... ./internal/...
GO_FOLDERS=$(shell echo ${GO_PACKAGES} | sed -e "s/\.\///g" | sed -e "s/\/\.\.\.//g")

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Compile the binary
	@mkdir -p bin
	@go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

mocks: ## generates mocks
	go install go.uber.org/mock/mockgen@v0.4.0
	go generate ./...

tests: ## runs all tests
	go test ./... -covermode=atomic

fmt: ## formats all go files
	go fmt ./...
	make format-lines

format-lines: ## formats all go files with golines
	go install github.com/segmentio/golines@latest
	golines -w -m 120 --ignore-generated --shorten-comments --ignored-dirs=${GO_LINES_IGNORED_DIRS} ${GO_FOLDERS}

lint: ## runs all linters
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

install: build ## compile the binary and copy it to PATH
	@sudo cp bin/$(APP_NAME) /usr/local/bin

build-linux-amd64: ## Compile the binary for amd64
	@env GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME)-linux-amd64 cmd/$(APP_NAME)/main.go

build-linux-arm64: ## Compile the binary for arm64
	@env GOOS=linux GOARCH=arm64 go build -o bin/$(APP_NAME)-linux-arm64 cmd/$(APP_NAME)/main.go

build-linux: ## Compile the binary for linux
	@env GOOS=linux go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

manifest: manifest-force ## Rebuild AVS specification manifest
	@./scripts/manifest.sh

manifest-force: ;
