.DEFAULT_GOAL := compile

export GOPATH            := $(abspath .)
export GOBIN             := $(GOPATH)/bin/$(shell uname -s | tr A-Z a-z)
export PATH              := $(GOBIN):$(PATH)

export AWS_DEFAULT_REGION ?= us-east-2

DOMAIN_NAME     ?= dev.kubernetes.deliver
COMPONENT_NAME  ?= git-service
REGISTRY        ?= $(subst https://,,$(lastword $(shell aws ecr get-login --region $(AWS_DEFAULT_REGION))))
IMAGE           ?= $(REGISTRY)/agilestacks/$(DOMAIN_NAME)/git-service
IMAGE_VERSION   ?= $(shell git rev-parse HEAD | colrm 7)
NAMESPACE       ?= automation-hub

BACKUP_BUCKET   ?= files.$(DOMAIN_NAME)
BACKUP_REGION   ?= us-east-2
REPOS_PATH      ?= /git
TS              := $(shell date +"%Y-%m-%d-%H-%M-%S")
BACKUP_SNAPSHOT ?= s3://$(BACKUP_BUCKET)/backup/$(DOMAIN_NAME)/git-service/$(COMPONENT_NAME)/$(TS).tar.bz

kubectl         ?= kubectl --context="$(DOMAIN_NAME)" --namespace="$(NAMESPACE)"
docker          ?= docker
aws             ?= aws

install:
	@go get -u github.com/mitchellh/gox
	@go get -u github.com/kardianos/govendor
.PHONY: install

govendor-list:
	@cd src/gits && $(GOBIN)/govendor list
.PHONY: govendor-list

govendor: govendor-list
	@cd src/gits && $(GOBIN)/govendor sync
.PHONY: govendor

govendor-add: govendor-list
	@cd src/gits && $(GOBIN)/govendor add +e
.PHONY: govendor-add

compile: govendor
	@$(GOBIN)/gox -rebuild \
		-osarch="darwin/amd64 linux/amd64" \
		-output=$(GOPATH)/bin/{{.OS}}/{{.Dir}} gits/...
.PHONY: compile

get:
	@go get gits
.PHONY: get

run: get
	$(GOBIN)/gits \
		-api_secret_env '' \
		-no_ext_api_calls \
		-repo_dir /tmp/gits \
		-aws_region us-east-2 \
		-aws_profile agilestacks \
		-trace
.PHONY: run

fmt:
	@go fmt gits gits/api gits/config gits/extapi gits/flags gits/repo gits/s3 gits/ssh
.PHONY: fmt

key:
	@ssh-keygen -t rsa -f gits-key -N ''
.PHONY: key

ifeq ($(RESTORE_SNAPSHOT),)
deploy: build ecr-login push kubernetes
else
deploy: build ecr-login push kubernetes wait restore
endif
.PHONY: deploy

# TODO replace with pod wait
wait:
	@echo Waiting for pod to come online
	@sleep 20
.PHONY: sleep

find-pod:
	@$(kubectl) get pods
	$(eval GITS_POD=$(shell $(kubectl) get pods -l project=git-service,provider=agilestacks.com,qualifier=gits --output=jsonpath={.items..metadata.name}))
.PHONY: find-pod

backup: find-pod
	$(kubectl) exec $(GITS_POD) -- /bin/sh -c \
		"tar cjf - -C $(REPOS_PATH) . | aws --region=$(BACKUP_REGION) s3 cp - $(BACKUP_SNAPSHOT)"
	@echo Outputs:
	@echo kind = git-service
	@echo component.git-service.snapshot = $(BACKUP_SNAPSHOT)
	@echo
.PHONY: backup

connect: find-pod
	$(kubectl) exec -ti $(GITS_POD) /bin/sh
.PHONY: connect

ifneq ($(RESTORE_SNAPSHOT),)
restore: find-pod
	$(kubectl) exec $(GITS_POD) -- /bin/sh -c \
		"aws s3 cp $(RESTORE_SNAPSHOT) - | tar xjpf - -C $(REPOS_PATH)"
.PHONY: restore
endif

delete-pod: find-pod
	$(kubectl) delete pod $(GITS_POD)
.PHONY: delete-pod

clean:
	-@rm -rf .cache pkg bin/darwin bin/linux \
		src/github.com src/golang.org src/gopkg.in \
		src/hub/vendor/github.com src/hub/vendor/golang.org src/hub/vendor/gopkg.in
	-@find src -not -path "*src/gits*" -not -path "src" -type d -maxdepth 1 | xargs rm -rf
.PHONY: clean

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

kubernetes:
	$(kubectl) apply -f templates/namespace.yaml
	$(kubectl) apply -f templates/volume.yaml
	$(kubectl) apply -f templates/pvc.yaml
	$(kubectl) apply -f templates/secret.yaml
	$(kubectl) apply -f templates/service.yaml
	$(kubectl) apply -f templates/ingress.yaml
	$(kubectl) apply -f templates/deployment.yaml
.PHONY: kubernetes

undeploy:
	-$(kubectl) delete -f templates/deployment.yaml
	-$(kubectl) delete -f templates/pvc.yaml
	-$(kubectl) delete -f templates/volume.yaml
	-$(kubectl) delete -f templates/ingress.yaml
	-$(kubectl) delete -f templates/service.yaml
	-$(kubectl) delete -f templates/secret.yaml
.PHONY: undeploy
