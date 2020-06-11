.DEFAULT_GOAL := get

OS := $(shell uname -s | tr A-Z a-z)

export GOBIN  := $(abspath .)/bin/$(OS)

export AWS_DEFAULT_REGION ?= us-east-2

DOMAIN_NAME    ?= test.dev.superhub.io
COMPONENT_NAME ?= git-service
LOCAL_IMAGE    ?= agilestacks/git-service
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

get:
	go get github.com/agilestacks/git-service/cmd/gits
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
	go fmt github.com/agilestacks/git-service/...
.PHONY: fmt

vet:
	go vet -composites=false github.com/agilestacks/git-service/...
.PHONY: vet

key:
	@ssh-keygen -t rsa -f gits-key -N ''
.PHONY: key

deploy: build ecr-login push kubernetes
ifneq ($(RESTORE_SNAPSHOT),)
deploy: wait restore
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
	@rm -f gits bin/gits
	@rm -rf bin/darwin bin/linux
.PHONY: clean

build:
	$(docker) build -t $(LOCAL_IMAGE):$(IMAGE_VERSION) -t $(LOCAL_IMAGE):latest .
.PHONY: build

ecr-login:
	$(aws) ecr get-login --no-include-email --region $(AWS_DEFAULT_REGION) | $(SHELL) -
.PHONY: ecr-login

push:
	$(docker) tag $(LOCAL_IMAGE):$(IMAGE_VERSION) $(IMAGE):$(IMAGE_VERSION)
	$(docker) tag $(LOCAL_IMAGE):$(IMAGE_VERSION) $(IMAGE):latest
	$(docker) push $(IMAGE):$(IMAGE_VERSION)
	$(docker) push $(IMAGE):latest
.PHONY: push

kubernetes:
	$(kubectl) apply -f templates/namespace.yaml
	$(kubectl) apply -f templates/pvc.yaml
	$(kubectl) apply -f templates/service.yaml
	$(kubectl) apply -f templates/$(INGRESS).yaml
	$(kubectl) apply -f templates/secret.yaml
	$(kubectl) apply -f templates/deployment.yaml
.PHONY: kubernetes

undeploy:
	-$(kubectl) delete -f templates/deployment.yaml
	-$(kubectl) delete -f templates/secret.yaml
	-$(kubectl) delete -f templates/$(INGRESS).yaml
	-$(kubectl) delete -f templates/service.yaml
	-$(kubectl) delete -f templates/pvc.yaml
.PHONY: undeploy
