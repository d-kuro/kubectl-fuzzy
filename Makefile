BINARY_NAME  = kubectl-fuzzy
LDFLAGS      = -ldflags="-s -w -X \"github.com/d-kuro/kubectl-fuzzy/pkg/cmd.Revision=$(shell git rev-parse --short HEAD)\""

export GO111MODULE=on

build:
	go build -o ./dist/$(BINARY_NAME) -v $(LDFLAGS) ./cmd/...
test:
	go test -race -covermode=atomic ./...
lint:
	golangci-lint run
install:
	go install ./cmd/kubectl-fuzzy/

