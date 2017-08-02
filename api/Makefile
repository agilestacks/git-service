.DEFAULT_GOAL := compile

export GOPATH            := $(abspath .)
export GOBIN             := $(GOPATH)/bin/$(shell uname -s | tr A-Z a-z)
export PATH              := $(GOBIN):$(PATH)

install:
	@go get -u github.com/mitchellh/gox
	@go get -u github.com/kardianos/govendor

govendor:
	@cd src/git && $(GOBIN)/govendor list
	@cd src/git && $(GOBIN)/govendor sync

compile: govendor
	@$(GOBIN)/gox -rebuild -osarch="darwin/amd64 linux/amd64" -output=$(GOPATH)/bin/{{.OS}}/{{.Dir}} hub/...

clean:
	@yes n | rm -rf .cache | true
	@yes n | rm -rf pkg | true
	@yes n | rm -rf bin/darwin | true
	@yes n | rm -rf linux/darwin | true
	@find src -not -path "*src/git*" -not -path "src" -type d -maxdepth 1 | xargs rm -rf
