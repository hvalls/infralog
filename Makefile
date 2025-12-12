.PHONY: build test lint docker-build docker-run docker-push docker-release help

BINARY_NAME := infralog
SRC_DIR := src
BUILD_DIR := bin
DOCKER_IMAGE := hvalls/infralog
DOCKER_TAG := latest

GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOVET := $(GOCMD) vet

.DEFAULT_GOAL := help

build:
	@mkdir -p $(BUILD_DIR)
	cd $(SRC_DIR) && $(GOBUILD) -o ../$(BUILD_DIR)/$(BINARY_NAME) -ldflags="-w -s" main.go
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

test:
	cd $(SRC_DIR) && $(GOTEST) ./...

test-coverage:
	cd $(SRC_DIR) && $(GOTEST) -coverprofile=coverage.out ./...
	cd $(SRC_DIR) && $(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: $(SRC_DIR)/coverage.html"

lint:
	cd $(SRC_DIR) && $(GOVET) ./...

fmt:
	cd $(SRC_DIR) && $(GOCMD) fmt ./...

deps:
	cd $(SRC_DIR) && $(GOMOD) download

deps-tidy:
	cd $(SRC_DIR) && $(GOMOD) tidy

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build
	docker run --rm \
		-p 8080:8080 \
		-v $(PWD)/examples/local-test/config-docker.yml:/etc/infralog/config.yml:ro \
		-v $(PWD)/examples/local-test:/data:ro \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-release:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

help:
	@echo "Infralog - Terraform State Change Monitor"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
