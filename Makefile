GO ?= go


lint: vet
	@golint -set_exit_status ./...
.PHONY: lint

test: lint
	@$(GO) test -v -race -short ./...
.PHONY: test

test-full: lint
	@$(GO) test -v -race ./...
.PHONY: test-full

vet:
	@$(GO) vet ./...
.PHONY: vet
