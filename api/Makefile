.DEFAULT_GOAL := compile

export GOPATH := $(abspath .)
export GOBIN  := $(GOPATH)/bin/$(shell uname -s | tr A-Z a-z)
export PATH   := $(GOBIN):$(PATH)

export AWS_DEFAULT_REGION ?= us-east-2

DOMAIN_NAME    ?= test.dev.superhub.io
COMPONENT_NAME ?= git-service
REGISTRY       ?= $(subst https://,,$(lastword $(shell aws ecr get-login --region $(AWS_DEFAULT_REGION))))
IMAGE          ?= $(REGISTRY)/agilestacks/$(DOMAIN_NAME)/git-service
IMAGE_VERSION  ?= $(shell git rev-parse HEAD | colrm 7)
NAMESPACE      ?= automation-hub
HUB_PROVIDES   ?=

ifneq (,$(filter tls-ingress,$(HUB_PROVIDES)))
	INGRESS:=ingress-tls
else
	INGRESS:=ingress
endif

BACKUP_BUCKET   ?= files.$(DOMAIN_NAME)
BACKUP_REGION   ?= us-east-2
REPOS_PATH      ?= /git
TS              := $(shell date +"%Y-%m-%d-%H-%M-%S")
BACKUP_SNAPSHOT ?= s3://$(BACKUP_BUCKET)/$(DOMAIN_NAME)/backup/git-service/$(COMPONENT_NAME)/$(TS).tar.bz

kubectl ?= kubectl --context="$(DOMAIN_NAME)" --namespace="$(NAMESPACE)"
docker  ?= docker
aws     ?= aws

running_git_pod := $(kubectl) get pods -l provider=agilestacks.com,project=git-service,qualifier=gits \
	-o jsonpath='{range.items[?(@.status.containerStatuses[0].ready==true)]}{.kind}{"\n"}{end}'

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

vet:
	@cd src && go vet -vettool=$(which shadow) \
		gits gits/api gits/config gits/extapi gits/flags gits/repo gits/s3 gits/ssh gits/util
.PHONY: vet

key:
	@ssh-keygen -t rsa -f gits-key -N ''
.PHONY: key

ifeq ($(RESTORE_SNAPSHOT),)
deploy: build ecr-login push kubernetes
else
deploy: build ecr-login push kubernetes wait restore
endif
.PHONY: deploy

wait:
	@echo "Waiting for Git Service pod to stand up"; \
	sleep 5; \
	for i in $$(seq 1 30); do \
		if test -n "$$($(running_git_pod))"; then \
			echo "done"; \
			exit 0; \
		fi; \
		echo "still waiting"; \
		sleep 10; \
	done; \
	echo "timeout"; \
	exit 1
.PHONY: wait

find-pod:
	@$(kubectl) get pods
	$(eval GITS_POD=$(shell $(kubectl) get pods -l project=git-service,provider=agilestacks.com,qualifier=gits --output=jsonpath={.items..metadata.name}))
.PHONY: find-pod

maintenance-on: find-pod
	$(kubectl) exec $(GITS_POD) -- \
		touch "$(REPOS_PATH)/_maintenance"
	@sleep 5
.PHONY: maintenance-on

maintenance-off: find-pod
	$(kubectl) exec $(GITS_POD) -- \
		rm -f "$(REPOS_PATH)/_maintenance"
.PHONY: maintenance-off

backup: maintenance-on tar maintenance-off

tar: find-pod
	$(kubectl) exec $(GITS_POD) -- /bin/sh -c \
		"tar cjf - -C $(REPOS_PATH) . | aws --region=$(BACKUP_REGION) s3 cp - $(BACKUP_SNAPSHOT)"
	@echo Outputs:
	@echo kind = git-service
	@echo component.git-service.snapshot = $(BACKUP_SNAPSHOT)
	@echo
.PHONY: tar

connect: find-pod
	$(kubectl) exec -ti $(GITS_POD) /bin/sh
.PHONY: connect

ifneq ($(RESTORE_SNAPSHOT),)
restore: maintenance-on untar maintenance-off

untar: find-pod
	$(kubectl) exec $(GITS_POD) -- /bin/sh -c \
		"aws s3 cp $(RESTORE_SNAPSHOT) - | tar xjpf - -C $(REPOS_PATH)"
.PHONY: untar
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
	$(kubectl) apply -f templates/pvc.yaml
	$(kubectl) apply -f templates/secret.yaml
	$(kubectl) apply -f templates/service.yaml
	$(kubectl) apply -f templates/$(INGRESS).yaml
	$(kubectl) apply -f templates/deployment.yaml
.PHONY: kubernetes

undeploy:
	-$(kubectl) delete -f templates/deployment.yaml
	-$(kubectl) delete -f templates/pvc.yaml
	-$(kubectl) delete -f templates/$(INGRESS).yaml
	-$(kubectl) delete -f templates/service.yaml
	-$(kubectl) delete -f templates/secret.yaml
.PHONY: undeploy
