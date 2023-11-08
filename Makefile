include .env

.DEFAULT_GOAL 	:= help
.PHONY: build help

build: ## Compile the binary
	@mkdir -p bin
	@go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'