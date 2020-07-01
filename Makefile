include *.mk
SHELL:=/bin/bash
REGISTRY=register.docker.com
IMAGE_TAG := $(shell git rev-parse --short HEAD)
IMAGE_NAME := kube-webhook


OS :=$(shell uname -s)

.PHONY: mod
mod:
	go mod download
	go mod vendor

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -cover -v `go list ./...`

.PHONY: mockgen
mockgen:
	mockgen -source=server/httpsrv/service.go -destination=server/httpsrv/mock/service.go
	mockgen -source=service/service.go -destination=service/mock/deps.go

.PHONY: svc
svc:
	go build -mod=vendor -o artifacts/svc ./cmd/svc

.PHONY: check-kubectl
check-kubectl:
	@which kubectl > /dev/null 2>&1 || (echo You\'re missing kubectl executable; @exit 1)

define k8s-deploy
	./make-cert.sh ${2} ${3}
	kubectl config use-context ${1} && \
	helm3 upgrade ${SERVICE} .helm \
	  --install --timeout 10m --wait \
	  --kube-context=${1} \
	  --namespace=${2} \
	  --set "global.namespace=${2}" \
	  --set "global.image=${REGISTRY}/${SERVICE}:${RELEASE}" \
	  --set "webhookName=k8s-acl.${2}.svc" \
	  --values ".helm/${4}" \
	  --values ".helm/cert_${2}_${3}.yaml"
	rm -f .helm/cert_${2}_${3}.yaml
endef

.PHONY: dockerise
dockerise:
	docker build -t ${IMAGE_NAME}:${IMAGE_TAG} -f Dockerfile .
	docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}

.PHONY: push_image_to_registry
push_image_to_registry:
	`AWS_SHARED_CREDENTIALS_FILE=~/.aws/credentials AWS_PROFILE=prof_name aws ecr get-login --region us-west-2 --no-include-email`
	docker push ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
	#docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:latest
	#docker push ${REGISTRY}/${IMAGE_NAME}:latest

# local
.PHONY: kube-local-deploy
kube-local-deploy:
	@echo Deploying release ${RELEASE} to local K8s Engine
	$(call k8s-deploy,${LOCAL_KUBE_CONTEXT},${LOCAL_NAMESPACE},${SERVICE},values-local.yaml)

.PHONY: local-cert
local-cert:
	./make-cert.sh ${LOCAL_NAMESPACE} ${SERVICE}

.PHONY: create-local-namespace
create-local-namespace:
	kubectl create namespace ${LOCAL_NAMESPACE}

.PHONY: deploy-helloapp-default
deploy-helloapp-default:
	helm upgrade hello-app ./test-apps/hello-app --install

.PHONY: deploy-helloapp-local
deploy-helloapp-local:
	helm upgrade hello-app ./test-apps/hello-app --install --namespace ${LOCAL_NAMESPACE}
