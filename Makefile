SHELL=/bin/bash
ROOT_DIR := $(shell pwd)
IMAGE_TAG := $(shell git rev-parse --short HEAD)
IMAGE_NAME := company/srv
REGISTRY := change-it.dkr.ecr.us-west-2.amazonaws.com

.PHONY: ci
ci: deps deps_check lint build test

.PHONY: mod
mod:
	GOSUMDB=off GO111MODULE=on GOPROXY=direct go mod download
	GOSUMDB=off GO111MODULE=on GOPROXY=direct go mod vendor

.PHONY: build
build:
	go build -o artifacts/svc

.PHONY: run
run:
	go run ./main.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -cover -v `go list ./...`

.PHONY: dockerise
dockerise:
	docker build -t ${IMAGE_NAME}:${IMAGE_TAG} -f Dockerfile .
	docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}

.PHONY: deploy
deploy:
	`AWS_SHARED_CREDENTIALS_FILE=~/.aws/credentials AWS_PROFILE=xid aws ecr get-login --region us-west-2 --no-include-email`
	docker push ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
	#docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:latest
	#docker push ${REGISTRY}/${IMAGE_NAME}:latest

.PHONY: mockgen
mockgen:
	#mockgen -source=service/service.go -destination=service/mock/deps.go
	mockgen -source=server/httpsrv/server.go -destination=server/httpsrv/mock/deps.go
