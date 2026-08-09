.PHONY: help check-docker env-dev env-prod build dev prod down logs ps image-name ecr-image-name ecr-build ecr-push

IMAGE_REGISTRY ?=
IMAGE_NAME ?= gin-react-admin-server
IMAGE_VERSION ?= $(shell cat VERSION)
ECR_REGISTRY ?= 622946927761.dkr.ecr.us-east-2.amazonaws.com
ECR_IMAGE_NAME ?= ai-api.ith5.cn
ECR_PLATFORM ?= linux/amd64
ECR_IMAGE := $(ECR_REGISTRY)/$(ECR_IMAGE_NAME):$(IMAGE_VERSION)
ECR_LATEST_IMAGE := $(ECR_REGISTRY)/$(ECR_IMAGE_NAME):latest
ALPINE_MIRROR ?= https://mirrors.aliyun.com/alpine
GOPROXY ?= https://goproxy.cn,direct
SERVER_IMAGE := $(if $(IMAGE_REGISTRY),$(IMAGE_REGISTRY)/,)$(IMAGE_NAME):$(IMAGE_VERSION)

help:
	@echo "Usage:"
	@echo "  make env-dev   Create .env.development from the example if missing"
	@echo "  make env-prod  Create .env.production from the example if missing"
	@echo "  make image-name Show the image tag that will be built/deployed"
	@echo "  make build     Build the versioned production server image"
	@echo "  make ecr-build Build the $(ECR_PLATFORM) server image for ECR"
	@echo "  make ecr-push  Build and push the versioned and latest ECR image tags"
	@echo "  make dev       Start the server with development env"
	@echo "  make prod      Build and start the server with production env"
	@echo "  make down      Stop server compose stacks"
	@echo "  make logs      Tail server logs"
	@echo "  make ps        Show server containers"

check-docker:
	@docker info >/dev/null 2>&1 || (echo "Docker is not running. Please start Docker Desktop or Docker Engine first."; exit 1)

env-dev:
	@test -f .env.development || cp .env.development.example .env.development

env-prod:
	@test -f .env.production || cp .env.production.example .env.production

image-name:
	@echo $(SERVER_IMAGE)

ecr-image-name:
	@echo $(ECR_IMAGE)

build: check-docker
	docker build \
		--build-arg ALPINE_MIRROR=$(ALPINE_MIRROR) \
		--build-arg GOPROXY=$(GOPROXY) \
		-t $(SERVER_IMAGE) .

ecr-build: check-docker
	docker buildx build \
		--platform $(ECR_PLATFORM) \
		--build-arg ALPINE_MIRROR=$(ALPINE_MIRROR) \
		--build-arg GOPROXY=$(GOPROXY) \
		-t $(ECR_IMAGE) \
		-t $(ECR_LATEST_IMAGE) \
		--load .

ecr-push: check-docker
	docker buildx build \
		--platform $(ECR_PLATFORM) \
		--build-arg ALPINE_MIRROR=$(ALPINE_MIRROR) \
		--build-arg GOPROXY=$(GOPROXY) \
		-t $(ECR_IMAGE) \
		-t $(ECR_LATEST_IMAGE) \
		--push .

dev: check-docker env-dev
	ALPINE_MIRROR=$(ALPINE_MIRROR) GOPROXY=$(GOPROXY) docker compose -f docker-compose.dev.yml up --build -d

prod: env-prod build
	SERVER_IMAGE=$(SERVER_IMAGE) docker compose -f docker-compose.prod.yml up -d

down: check-docker
	docker compose -f docker-compose.dev.yml down
	docker compose -f docker-compose.prod.yml down

logs: check-docker
	docker compose -f docker-compose.dev.yml logs -f server

ps: check-docker
	docker compose -f docker-compose.dev.yml ps
	docker compose -f docker-compose.prod.yml ps
