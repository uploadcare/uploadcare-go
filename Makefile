all: lint vet test

lint:
	golint -set_exit_status ./...
.PHONY: lint

test:
	go test -v -race -short ./...
.PHONY: test

test-full:
	go test -v -race ./...
.PHONY: test-full

vet:
	go vet ./...
.PHONY: vet
