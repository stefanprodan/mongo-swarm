SHELL:=/bin/bash

APP_VERSION?=0.1.0

DIST:=$$(pwd)/dist
BUILD_DATE:=$(shell date -u +%Y-%m-%d_%H.%M.%S)

REGISTRY?=index.docker.io
REPOSITORY?=stefanprodan

build:
	@echo ">>> Building mongo-loadtest image"
	@docker build -t mongo-loadtest:$(APP_VERSION) \
		--build-arg APP_VERSION=$(APP_VERSION) .

push:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"
	@echo ">>> Pushing loadtest to $(REGISTRY)/$(REPOSITORY)"
	@docker tag mongo-loadtest:$(APP_VERSION) $(REPOSITORY)/mongo-loadtest:$(APP_VERSION)
	@docker tag mongo-loadtest:$(APP_VERSION) $(REPOSITORY)/mongo-loadtest:latest
	@docker push $(REPOSITORY)/mongo-loadtest:$(APP_VERSION)
	@docker push $(REPOSITORY)/mongo-loadtest:latest
