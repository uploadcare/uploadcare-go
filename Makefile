all: lint vet test

lint:
	golint -set_exit_status ./...
.PHONY: lint

test:
	go test -v -race ./...
.PHONY: test

vet:
	go vet ./...
.PHONY: vet
