.DEFAULT_GOAL := compile

export GOPATH            := $(abspath .)
export GOBIN             := $(GOPATH)/bin/$(shell uname -s | tr A-Z a-z)
export PATH              := $(GOBIN):$(PATH)

export TF_VAR_name        ?= dev
export TF_VAR_base_domain ?= stack.delivery
export AWS_DEFAULT_REGION ?= us-east-2
export REGISTRY           ?= $(subst https://,,$(lastword $(shell aws ecr get-login --region $(AWS_DEFAULT_REGION))))
export IMAGE              ?= $(REGISTRY)/agilestacks/$(TF_VAR_name)/git-service
export IMAGE_VERSION      ?= $(shell git rev-parse HEAD | colrm 7)
export NAMESPACE          ?= automation-hub
export kubectl            ?= kubectl --context="$(TF_VAR_name).$(TF_VAR_base_domain)" --namespace="$(NAMESPACE)"

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

run:
	go get gits && $(GOBIN)/gits \
		-api_secret_env '' \
		-no_ext_api_calls \
		-repo_dir /tmp/gits \
		-aws_region us-east-2 \
		-aws_profile agilestacks \
		-debug

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

build: govendor
	@docker run --rm \
		-v "$(GOPATH)/src/gits":/usr/local/go/src/gits \
		-v "$(GOPATH)/bin/linux":/linuxbuild \
		-u $(shell id -u) \
		golang:1.8-alpine go build -o /linuxbuild/gits-musl -v gits
	@docker build -t $(IMAGE):$(IMAGE_VERSION) .
.PHONY: build

push:
	aws ecr get-login --region $(AWS_DEFAULT_REGION) | sed -e 's/[ +]-e[ +]none[ +]/ /g' | sh -
	docker tag  $(IMAGE):$(IMAGE_VERSION) $(IMAGE):latest
	docker push $(IMAGE):$(IMAGE_VERSION)
	docker push $(IMAGE):latest
.PHONY: push
	
deploy: install build push
	$(kubectl) apply -f templates/namespace.yaml
	$(kubectl) apply -f templates/volume.yaml	
	$(kubectl) apply -f templates/pvc.yaml	
	$(kubectl) apply -f templates/secret.yaml	
	$(kubectl) apply -f templates/service.yaml
	$(kubectl) apply -f templates/deployment.yaml
.PHONY: deploy

undeploy:
	$(kubectl) delete -f templates/deployment.yaml | true
	$(kubectl) apply -f templates/pvc.yaml	       | true
	$(kubectl) apply -f templates/volume.yaml	     | true
	$(kubectl) delete -f templates/service.yaml    | true
	$(kubectl) delete -f templates/secret.yaml     | true	
.PHONY: undeploy
