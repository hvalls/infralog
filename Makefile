.PHONY: build run test lint clean docker-build docker-run docker-compose-up help

# Variables
BINARY_NAME := infralog
SRC_DIR := src
BUILD_DIR := bin
DOCKER_IMAGE := infralog
DOCKER_TAG := latest

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOVET := $(GOCMD) vet

# Default target
.DEFAULT_GOAL := help

## Build

build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	cd $(SRC_DIR) && $(GOBUILD) -o ../$(BUILD_DIR)/$(BINARY_NAME) -ldflags="-w -s" main.go
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

build-dev: ## Build the binary without optimizations (faster compilation)
	@mkdir -p $(BUILD_DIR)
	cd $(SRC_DIR) && $(GOBUILD) -o ../$(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

## Run

run: build ## Build and run with example config
	./$(BUILD_DIR)/$(BINARY_NAME) --config-file examples/local-test/config.yml

run-dev: ## Run without building (using go run)
	cd $(SRC_DIR) && $(GOCMD) run main.go --config-file ../examples/local-test/config.yml

## Test

test: ## Run all tests
	cd $(SRC_DIR) && $(GOTEST) ./...

test-verbose: ## Run all tests with verbose output
	cd $(SRC_DIR) && $(GOTEST) -v ./...

test-coverage: ## Run tests with coverage report
	cd $(SRC_DIR) && $(GOTEST) -coverprofile=coverage.out ./...
	cd $(SRC_DIR) && $(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: $(SRC_DIR)/coverage.html"

## Lint & Format

lint: ## Run go vet
	cd $(SRC_DIR) && $(GOVET) ./...

fmt: ## Format code
	cd $(SRC_DIR) && $(GOCMD) fmt ./...

## Dependencies

deps: ## Download dependencies
	cd $(SRC_DIR) && $(GOMOD) download

deps-tidy: ## Tidy dependencies
	cd $(SRC_DIR) && $(GOMOD) tidy

## Docker

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build ## Build and run Docker container with local test config
	docker run --rm \
		-v $(PWD)/examples/local-test/config-docker.yml:/etc/infralog/config.yml:ro \
		-v $(PWD)/examples/local-test:/data:ro \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up: ## Run with docker-compose
	docker-compose up --build

docker-compose-down: ## Stop docker-compose
	docker-compose down

## Local Testing

simulate: ## Run the state change simulator
	./examples/local-test/simulate-changes.sh

init-test: ## Initialize test state file
	cp examples/local-test/state_v1.json examples/local-test/terraform.tfstate

## Clean

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
	rm -f $(SRC_DIR)/coverage.out $(SRC_DIR)/coverage.html

clean-docker: ## Remove Docker image
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true

clean-all: clean clean-docker ## Remove all artifacts

## Help

help: ## Show this help
	@echo "Infralog - Terraform State Change Monitor"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
