.DEFAULT_GOAL := compile

export GOPATH            := $(abspath .)
export GOBIN             := $(GOPATH)/bin/$(shell uname -s | tr A-Z a-z)
export PATH              := $(GOBIN):$(PATH)

export AWS_ACCOUNT_NUMBER ?= 973998981304
export TF_VAR_name        ?= dev
export AWS_DEFAULT_REGION ?= us-east-2
export IMAGE              ?= $(AWS_ACCOUNT_NUMBER).dkr.ecr.$(AWS_DEFAULT_REGION).amazonaws.com/agilestacks/$(TF_VAR_name)/git-service
export IMAGE_NAME         ?= $(IMAGE):$(IMAGE_VERSION)
export IMAGE_VERSION      ?= $(shell git rev-parse HEAD | colrm 7)
export IMAGE_NAME_LATEST  ?= $(IMAGE):latest

export DOMAIN_NAME        ?= $(TF_VAR_name).$(TF_VAR_base_domain)
export NAMESPACE          ?= automation-hub
export kubectl            ?= kubectl --context="$(DOMAIN_NAME)" --namespace="$(NAMESPACE)"

install:
	@go get -u github.com/mitchellh/gox
	@go get -u github.com/kardianos/govendor

govendor:
	@cd src/gits && $(GOBIN)/govendor list
	@cd src/gits && $(GOBIN)/govendor sync

compile: govendor
	@$(GOBIN)/gox -rebuild -osarch="darwin/amd64 linux/amd64" -output=$(GOPATH)/bin/{{.OS}}/{{.Dir}} gits/...

fmt:
	@go fmt gits gits/api gits/config gits/repo gits/s3 gits/ssh

key:
	@ssh-keygen -t rsa -f gits-key -N ''

clean:
	@yes n | rm -rf .cache | true
	@yes n | rm -rf pkg | true
	@yes n | rm -rf bin/darwin | true
	@yes n | rm -rf linux/darwin | true
	@find src -not -path "*src/gits*" -not -path "src" -type d -maxdepth 1 | xargs rm -rf

build:
	@docker build -t $(IMAGE_NAME) .
.PHONY: build

push:
	aws ecr get-login --region $(AWS_DEFAULT_REGION) | sed -e 's/[ +]-e[ +]none[ +]/ /g' | sh -
	docker tag  $(IMAGE_NAME) $(IMAGE_NAME_LATEST)
	docker push $(IMAGE_NAME)
	docker push $(IMAGE_NAME_LATEST)
.PHONY: push

deploy: install compile build push
	# $(kubectl) apply -f templates/namespace.yaml
	# $(kubectl) apply -f templates/service.yaml
	# $(kubectl) apply -f templates/deployment.yaml
.PHONY: deploy

undeploy:
	$(kubectl) delete -f templates/deployment.yaml | true
	$(kubectl) delete -f templates/service.yaml    | true
	$(kubectl) delete -f templates/namespace.yaml  | true
.PHONY: undeploy