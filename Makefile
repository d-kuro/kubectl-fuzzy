BINARY_NAME  = kubectl-fzf
LDFLAGS      = -ldflags="-s -w -X \"github.com/d-kuro/kubectl-fzf/pkg/cmd.Revision=$(shell git rev-parse --short HEAD)\""

export GO111MODULE=on

build:
	go build -o ./dist/$(BINARY_NAME) -v $(LDFLAGS) ./cmd/...
test:
	go test -race -covermode=atomic ./...
lint:
	golangci-lint run
