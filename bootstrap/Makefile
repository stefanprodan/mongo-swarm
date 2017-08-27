SHELL:=/bin/bash

APP_VERSION?=1.0.0

DIST:=$$(pwd)/dist
BUILD_DATE:=$(shell date -u +%Y-%m-%d_%H.%M.%S)

REGISTRY?=index.docker.io
REPOSITORY?=stefanprodan

build:
	@echo ">>> Building bootstrap-build image"
	@docker build -t bootstrap-build:$(BUILD_DATE) -f build.dockerfile .

	@echo ">>> Building bootstrap"
	@docker run --rm  -v "$(DIST):/go/dist" bootstrap-build:$(BUILD_DATE) \
		go build -ldflags "-X main.version=$(APP_VERSION)" -o /go/dist/bootstrap github.com/stefanprodan/mongo-swarm/bootstrap

	@docker rmi bootstrap-build:$(BUILD_DATE)

	@echo ">>> Build artifacts:"
	@find dist -type f -print0 | xargs -0 ls -t

pack:
	@echo ">>> Building mongo-bootstrap image for deploy"
	@docker build -t mongo-bootstrap:$(APP_VERSION) \
		-f deploy.dockerfile .

	@rm -rf dist

push:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"
	@echo ">>> Pushing bootstrap to $(REGISTRY)/$(REPOSITORY)"
	@docker tag mongo-bootstrap:$(APP_VERSION) $(REPOSITORY)/mongo-bootstrap:$(APP_VERSION)
	@docker tag mongo-bootstrap:$(APP_VERSION) $(REPOSITORY)/mongo-bootstrap:latest
	@docker push $(REPOSITORY)/mongo-bootstrap:$(APP_VERSION)
	@docker push $(REPOSITORY)/mongo-bootstrap:latest
