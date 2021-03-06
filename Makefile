PKG        := ./...
GOCC       ?= $(shell command -v "go")
GOFMT      ?= $(shell command -v "gofmt")
GO         ?= GOPATH=$(GOPATH) $(GOCC)

.PHONY: all
all: format build test check

.PHONY: format
format: $(GOCC)
	find . -iname '*.go' | xargs -n1 $(GOFMT) -w -s

.PHONY: build
build:
	$(GO) build $(GOFLAGS) -v

.PHONY: test
test:
	$(GO) test $(GOFLAGS) -i $(PKG)

.PHONY: check
check:
	gometalinter $(PKG) --concurrency=1 --deadline=5m
