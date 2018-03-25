.DEFAULT_GOAL := compile

export GOPATH            := $(abspath .)
export GOBIN             := $(GOPATH)/bin/$(shell uname -s | tr A-Z a-z)
export PATH              := $(GOBIN):$(PATH)

export AWS_DEFAULT_REGION ?= us-east-2

NAME          ?= dev
DOMAIN_NAME   ?= stack.delivery
REGISTRY      ?= $(subst https://,,$(lastword $(shell aws ecr get-login --region $(AWS_DEFAULT_REGION))))
IMAGE         ?= $(REGISTRY)/agilestacks/$(DOMAIN_NAME)/git-service
IMAGE_VERSION ?= $(shell git rev-parse HEAD | colrm 7)
NAMESPACE     ?= automation-hub
kubectl       ?= kubectl --context="$(DOMAIN_NAME)" --namespace="$(NAMESPACE)"
docker        ?= docker
aws           ?= aws

install:
	@go get -u github.com/mitchellh/gox
	@go get -u github.com/kardianos/govendor

govendor:
	@cd src/gits && $(GOBIN)/govendor list
	@cd src/gits && $(GOBIN)/govendor sync

compile: govendor
	@$(GOBIN)/gox -rebuild \
		-osarch="darwin/amd64 linux/amd64" \
		-output=$(GOPATH)/bin/{{.OS}}/{{.Dir}} gits/...

get:
	@go get gits

run:
	go get gits && $(GOBIN)/gits \
		-api_secret_env '' \
		-no_ext_api_calls \
		-repo_dir /tmp/gits \
		-aws_region us-east-2 \
		-aws_profile agilestacks \
		-trace

fmt:
	@go fmt gits gits/api gits/config gits/extapi gits/flags gits/repo gits/s3 gits/ssh

key:
	@ssh-keygen -t rsa -f gits-key -N ''

clean:
	@yes n | rm -rf .cache | true
	@yes n | rm -rf pkg | true
	@yes n | rm -rf bin/darwin | true
	@yes n | rm -rf linux/darwin | true
	@find src -not -path "*src/gits*" -not -path "src" -type d -maxdepth 1 | xargs rm -rf

build:
	@$(docker) build -t $(IMAGE):$(IMAGE_VERSION) .
.PHONY: build

ecr-login:
	$(aws) ecr get-login --no-include-email --region $(AWS_DEFAULT_REGION) | $(SHELL) -
.PHONY: ecr-login

push:
	$(docker) tag  $(IMAGE):$(IMAGE_VERSION) $(IMAGE):latest
	$(docker) push $(IMAGE):$(IMAGE_VERSION)
	$(docker) push $(IMAGE):latest
.PHONY: push

deploy: build ecr-login push
	$(kubectl) apply -f templates/namespace.yaml
	$(kubectl) apply -f templates/volume.yaml
	$(kubectl) apply -f templates/pvc.yaml
	$(kubectl) apply -f templates/secret.yaml
	$(kubectl) apply -f templates/service.yaml
	$(kubectl) apply -f templates/ingress.yaml
	$(kubectl) apply -f templates/deployment.yaml
.PHONY: deploy

undeploy:
	$(kubectl) delete -f templates/deployment.yaml | true
	$(kubectl) apply -f templates/pvc.yaml         | true
	$(kubectl) apply -f templates/volume.yaml      | true
	$(kubectl) delete -f templates/ingress.yaml    | true
	$(kubectl) delete -f templates/service.yaml    | true
	$(kubectl) delete -f templates/secret.yaml     | true
.PHONY: undeploy
