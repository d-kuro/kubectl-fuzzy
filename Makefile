BINARY_NAME  = kubectl-fuzzy
LDFLAGS      = -ldflags="-s -w -X \"github.com/d-kuro/kubectl-fuzzy/pkg/cmd.Revision=$(shell git rev-parse --short HEAD)\""

export GO111MODULE=on

.PHONY: build
build:
	go build -o ./dist/$(BINARY_NAME) -v $(LDFLAGS) ./cmd/...

.PHONY: test
test:
	go test -race -covermode=atomic ./...

.PHONY: lint
lint:
	golangci-lint run --fix

.PHONY: install
install:
	go install ./cmd/kubectl-fuzzy/
