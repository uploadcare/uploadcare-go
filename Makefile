GO ?= go
GOLANGCI_LINT ?= golangci-lint

lint:
	@$(GOLANGCI_LINT) run ./...
.PHONY: lint

vet:
	@$(GO) vet ./...
.PHONY: vet

test: lint vet
	@$(GO) test -v -race -short ./...
.PHONY: test

test-full: vet
	@$(GO) test -v -race ./...
.PHONY: test-full
