GO ?= go
GOLANGCI_LINT ?= golangci-lint

vet:
	@$(GO) vet ./...
.PHONY: vet

lint:
	@$(GOLANGCI_LINT) run ./...
.PHONY: lint

test: vet
	@$(GO) test -v -race -short ./...
.PHONY: test

test-full: vet
	@$(GO) test -v -race ./...
.PHONY: test-full
