SHELL:=/bin/bash

APP_VERSION?=0.4.0

DIST:=$$(pwd)/dist
BUILD_DATE:=$(shell date -u +%Y-%m-%d_%H.%M.%S)

REGISTRY?=index.docker.io
REPOSITORY?=stefanprodan

build:
	@echo ">>> Building mongo-bootstrap image"
	@docker build -t mongo-bootstrap:$(APP_VERSION) \
		--build-arg APP_VERSION=$(APP_VERSION) .

push:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"
	@echo ">>> Pushing bootstrap to $(REGISTRY)/$(REPOSITORY)"
	@docker tag mongo-bootstrap:$(APP_VERSION) $(REPOSITORY)/mongo-bootstrap:$(APP_VERSION)
	@docker tag mongo-bootstrap:$(APP_VERSION) $(REPOSITORY)/mongo-bootstrap:latest
	@docker push $(REPOSITORY)/mongo-bootstrap:$(APP_VERSION)
	@docker push $(REPOSITORY)/mongo-bootstrap:latest
