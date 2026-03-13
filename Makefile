GO ?= go

vet:
	@$(GO) vet ./...
.PHONY: vet

test: vet
	@$(GO) test -v -race -short ./...
.PHONY: test

test-full: vet
	@$(GO) test -v -race ./...
.PHONY: test-full
