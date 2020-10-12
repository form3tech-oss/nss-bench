SHELL := /bin/bash

ROOT := $(shell git rev-parse --show-toplevel)

VERSION ?= $(shell git describe --always --dirty="-dev")

DOCKER_IMG ?= form3tech/nss-bench
DOCKER_TAG ?= $(VERSION)

.PHONY: docker.build
docker.build:
	docker build -t $(DOCKER_IMG):$(DOCKER_TAG) $(ROOT)

.PHONY: docker.push
docker.push: docker.build
	docker push $(DOCKER_IMG):$(DOCKER_TAG)
